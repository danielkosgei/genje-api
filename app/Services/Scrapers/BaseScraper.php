<?php

namespace App\Services\Scrapers;

use App\Models\News;
use App\Services\NewsScraperInterface;
use App\Services\RssFeedService;
use Illuminate\Support\Facades\Log;

abstract class BaseScraper implements NewsScraperInterface
{
    protected RssFeedService $rssService;

    public function __construct(RssFeedService $rssService)
    {
        $this->rssService = $rssService;
    }

    /**
     * Get RSS feed URL(s) for this source
     */
    abstract protected function getFeedUrls(): array;

    /**
     * Extract article data from RSS item
     */
    abstract protected function extractArticleData($item): ?array;

    /**
     * Get the source name
     */
    abstract public function getSourceName(): string;

    /**
     * Scrape news articles
     */
    public function scrape(): int
    {
        $count = 0;
        $feedUrls = $this->getFeedUrls();

        foreach ($feedUrls as $feedUrl) {
            $xml = $this->rssService->fetchFeed($feedUrl);

            if (!$xml) {
                continue;
            }

            $items = $this->rssService->extractItems($xml);

            foreach ($items as $item) {
                $articleData = $this->extractArticleData($item);

                if (!$articleData) {
                    continue;
                }

                // Check if article already exists
                if (News::where('url', $articleData['url'])->exists()) {
                    continue;
                }

                try {
                    News::create($articleData);
                    $count++;
                } catch (\Exception $e) {
                    Log::error("Failed to save article: {$articleData['url']}", [
                        'error' => $e->getMessage(),
                        'source' => $this->getSourceName(),
                    ]);
                }
            }
        }

        Log::info("Scraped {$count} articles from {$this->getSourceName()}");
        return $count;
    }

    /**
     * Extract image URL from content or media tags
     */
    protected function extractImageUrl($item): ?string
    {
        // Try media:content or media:thumbnail (RSS with media namespace)
        $namespaces = $item->getNamespaces(true);
        
        if (isset($namespaces['media'])) {
            $mediaContent = $item->xpath('.//media:content');
            if ($mediaContent && count($mediaContent) > 0) {
                $attrs = $mediaContent[0]->attributes();
                if (isset($attrs['url'])) {
                    return trim((string) $attrs['url']);
                }
            }

            $mediaThumbnail = $item->xpath('.//media:thumbnail');
            if ($mediaThumbnail && count($mediaThumbnail) > 0) {
                $attrs = $mediaThumbnail[0]->attributes();
                if (isset($attrs['url'])) {
                    return trim((string) $attrs['url']);
                }
            }
        }

        // Try enclosure
        if (isset($item->enclosure)) {
            $enclosure = $item->enclosure;
            $attrs = $enclosure->attributes();
            if (isset($attrs['url'])) {
                $url = trim((string) $attrs['url']);
                if (preg_match('/\.(jpg|jpeg|png|gif|webp)$/i', $url)) {
                    return $url;
                }
            }
        }

        // Try to extract from description/content
        $description = $this->rssService->getText($item, './/description');
        if ($description) {
            preg_match('/<img[^>]+src=["\']([^"\']+)["\']/i', $description, $matches);
            if (isset($matches[1])) {
                return $matches[1];
            }
        }

        return null;
    }

    /**
     * Clean HTML from text
     */
    protected function cleanHtml(string $text): string
    {
        $text = strip_tags($text);
        $text = html_entity_decode($text, ENT_QUOTES | ENT_HTML5, 'UTF-8');
        $text = trim($text);
        return $text;
    }
}

