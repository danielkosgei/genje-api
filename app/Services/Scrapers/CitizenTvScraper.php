<?php

namespace App\Services\Scrapers;

use App\Models\News;
use Carbon\Carbon;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;
use Symfony\Component\DomCrawler\Crawler;

class CitizenTvScraper extends BaseScraper
{
    private const SITEMAP_URL = 'https://citizen.digital/sitemap.xml';
    private const NEWS_NAMESPACE = 'http://www.google.com/schemas/sitemap-news/0.9';

    public function getSourceName(): string
    {
        return 'Citizen TV';
    }

    /**
     * Override the generic RSS workflow and ingest Citizen Digital directly from
     * their news sitemap to avoid Google News thumbnails and redirections.
     */
    public function scrape(): int
    {
        $count = 0;

        try {
            $response = Http::timeout(30)
                ->withHeaders([
                    'User-Agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
                    'Accept' => 'application/xml,text/xml;q=0.9,*/*;q=0.8',
                ])->get(self::SITEMAP_URL);

            if (!$response->successful()) {
                Log::warning('Citizen sitemap fetch failed', ['status' => $response->status()]);
                return 0;
            }

            $xml = @simplexml_load_string($response->body());
            if (!$xml) {
                Log::warning('Citizen sitemap XML parse failed');
                return 0;
            }

            foreach ($xml->url as $urlNode) {
                $loc = (string) $urlNode->loc;
                if ($loc === '' || !Str::contains($loc, '/article/')) {
                    continue;
                }

                if (News::where('url', $loc)->exists()) {
                    continue;
                }

                $newsData = $urlNode->children(self::NEWS_NAMESPACE)->news ?? null;
                $publishedAt = $this->parsePublicationDate((string) ($newsData->publication_date ?? ''));

                if ($publishedAt && $publishedAt->lt(now()->subDays(2))) {
                    continue;
                }

                $articleData = $this->fetchArticleMetadata($loc, $publishedAt);
                if (!$articleData) {
                    continue;
                }

                $articleData['fingerprint'] = $this->computeFingerprint(
                    $articleData['source'] ?? '',
                    $articleData['title'] ?? '',
                    $articleData['published_at'] ?? null
                );

                if (
                    !empty($articleData['fingerprint']) &&
                    News::where('fingerprint', $articleData['fingerprint'])->exists()
                ) {
                    continue;
                }

                $articleData['quality_score'] = $this->computeQualityScore(
                    $articleData['title'] ?? '',
                    $articleData['description'] ?? '',
                    (bool) ($articleData['image_url'] ?? null),
                    $articleData['published_at'] ?? null
                );

                try {
                    $created = News::create($articleData);
                    if (!empty($created->image_url)) {
                        \App\Jobs\CacheArticleImageJob::dispatch($created->id);
                    } else {
                        \App\Jobs\BackfillArticleImageJob::dispatch($created->id);
                    }
                    $count++;
                } catch (\Exception $e) {
                    Log::error('Failed to persist Citizen article', [
                        'url' => $loc,
                        'error' => $e->getMessage(),
                    ]);
                }
            }
        } catch (\Throwable $e) {
            Log::error('Citizen sitemap ingestion failed', ['error' => $e->getMessage()]);
        }

        Log::info("Scraped {$count} articles from {$this->getSourceName()}");
        return $count;
    }

    private function fetchArticleMetadata(string $url, ?Carbon $fallbackPublishedAt): ?array
    {
        try {
            $response = Http::timeout(30)
                ->withHeaders([
                    'User-Agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
                    'Accept' => 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
                ])->get($url);

            if (!$response->successful()) {
                return null;
            }

            $html = $response->body();
            $crawler = new Crawler($html, $url);

            $title = $this->firstMeta($crawler, 'property', 'og:title')
                ?? $this->firstMeta($crawler, 'name', 'twitter:title');
            $description = $this->firstMeta($crawler, 'property', 'og:description')
                ?? $this->firstMeta($crawler, 'name', 'description');
            $image = $this->sanitizeImageUrl($this->firstMeta($crawler, 'property', 'og:image'));
            $author = $this->firstMeta($crawler, 'name', 'author')
                ?? $this->firstMeta($crawler, 'property', 'article:author');
            $category = $this->firstMeta($crawler, 'property', 'article:section');

            $published = $this->parsePublicationDate(
                $this->firstMeta($crawler, 'property', 'article:published_time')
            ) ?? $fallbackPublishedAt ?? now();

            if (!$title) {
                return null;
            }

            $cleanDescription = $description ? $this->cleanHtml($description) : null;

            return [
                'title' => $this->cleanHtml($title),
                'description' => $cleanDescription,
                'content' => $cleanDescription,
                'source' => $this->getSourceName(),
                'category' => $category ? Str::lower($this->cleanHtml($category)) : null,
                'url' => $url,
                'image_url' => $image,
                'author' => $author ? $this->cleanHtml($author) : null,
                'published_at' => $published,
            ];
        } catch (\Throwable $e) {
            Log::debug('Citizen article fetch failed', [
                'url' => $url,
                'error' => $e->getMessage(),
            ]);
            return null;
        }
    }

    private function firstMeta(Crawler $crawler, string $attribute, string $value): ?string
    {
        try {
            return trim((string) $crawler
                ->filter("meta[{$attribute}='{$value}']")
                ->first()
                ->attr('content'));
        } catch (\Throwable $e) {
            return null;
        }
    }

    private function parsePublicationDate(?string $value): ?Carbon
    {
        if (!$value) {
            return null;
        }

        try {
            return Carbon::parse($value);
        } catch (\Throwable $e) {
            return null;
        }
    }

    /**
     * Unused in direct sitemap ingestion but required by the abstract base class.
     */
    protected function getFeedUrls(): array
    {
        return [];
    }

    protected function extractArticleData($item): ?array
    {
        return null;
    }
}

