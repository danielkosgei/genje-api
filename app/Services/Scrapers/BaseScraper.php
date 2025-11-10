<?php

namespace App\Services\Scrapers;

use App\Models\News;
use App\Services\NewsScraperInterface;
use App\Services\RssFeedService;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;

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

            $items = array_slice($this->rssService->extractItems($xml), 0, 25);

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
     * Decode description HTML for parsing
     */
    protected function getDecodedDescription($item): ?string
    {
        $description = $this->rssService->getText($item, './/description');

        if ($description === '') {
            return null;
        }

        return html_entity_decode($description, ENT_QUOTES | ENT_HTML5, 'UTF-8');
    }

    /**
     * Extract article link from common RSS fields
     */
    protected function extractLinkFromItem($item): ?string
    {
        $link = $this->rssService->getText($item, './/link');

        if ($link !== '') {
            return $link;
        }

        $decodedDescription = $this->getDecodedDescription($item);
        if ($decodedDescription && preg_match('/href=["\']([^"\']+)["\']/', $decodedDescription, $matches)) {
            return trim($matches[1]);
        }

        $guid = $this->rssService->getText($item, './/guid');
        if ($guid !== '') {
            return $guid;
        }

        return null;
    }

    /**
     * Extract category text if present
     */
    protected function extractCategoryFromItem($item): ?string
    {
        $category = $this->rssService->getText($item, './/category');

        if ($category === '') {
            return null;
        }

        return Str::lower($this->cleanHtml($category));
    }

    /**
     * Extract author/source text if present
     */
    protected function extractAuthorFromItem($item): ?string
    {
        $author = $this->rssService->getText($item, './/author');
        if ($author !== '') {
            return $this->cleanHtml($author);
        }

        $source = $this->rssService->getText($item, './/source');
        if ($source !== '') {
            return $this->cleanHtml($source);
        }

        return null;
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
        $description = $this->getDecodedDescription($item);
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
    protected function cleanHtml(?string $text): string
    {
        if ($text === null) {
            return '';
        }

        $text = strip_tags($text);
        $text = html_entity_decode($text, ENT_QUOTES | ENT_HTML5, 'UTF-8');
        $text = trim($text);
        return $text;
    }
}

