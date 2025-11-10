<?php

namespace App\Services\Scrapers;

use App\Models\News;
use Carbon\Carbon;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;
use Symfony\Component\DomCrawler\Crawler;

class StarScraper extends BaseScraper
{
    private const INDEX_URL = 'https://www.the-star.co.ke/news';

    public function getSourceName(): string
    {
        return 'The Star';
    }

    /**
     * Scrape The Star directly from their news index and individual articles.
     */
    public function scrape(): int
    {
        $count = 0;

        try {
            $response = Http::timeout(30)
                ->withHeaders([
                    'User-Agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
                    'Accept' => 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
                ])->get(self::INDEX_URL);

            if (!$response->successful()) {
                Log::warning('The Star index fetch failed', ['status' => $response->status()]);
                return 0;
            }

            $crawler = new Crawler($response->body(), self::INDEX_URL);
            $articleUrls = array_slice($this->extractArticleUrls($crawler), 0, 20);

            foreach ($articleUrls as $url) {
                if (News::where('url', $url)->exists()) {
                    continue;
                }

                $articleData = $this->fetchArticleMetadata($url);
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
                    Log::error('Failed to persist The Star article', [
                        'url' => $url,
                        'error' => $e->getMessage(),
                    ]);
                }
            }
        } catch (\Throwable $e) {
            Log::error('The Star ingestion failed', ['error' => $e->getMessage()]);
        }

        Log::info("Scraped {$count} articles from {$this->getSourceName()}");
        return $count;
    }

    private function extractArticleUrls(Crawler $crawler): array
    {
        $urls = [];

        $crawler->filter('div.flex.group')->each(function (Crawler $node) use (&$urls) {
            try {
                $titleNode = $node->filter('h6')->first();
                if (!$titleNode->count()) {
                    return;
                }

                $linkNode = $titleNode->ancestors()->filter('a[href^="/news/"]')->first();
                if (!$linkNode->count()) {
                    return;
                }

                $href = trim((string) $linkNode->attr('href'));
                if ($href === '') {
                    return;
                }

                $urls[] = Str::startsWith($href, 'http')
                    ? $href
                    : 'https://www.the-star.co.ke' . $href;
            } catch (\Throwable $e) {
                // Ignore parsing errors for individual nodes
            }
        });

        return array_values(array_unique($urls));
    }

    private function fetchArticleMetadata(string $url): ?array
    {
        try {
            $response = Http::timeout(30)
                ->withHeaders([
                    'User-Agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
                    'Accept' => 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
                    'Referer' => self::INDEX_URL,
                ])->get($url);

            if (!$response->successful()) {
                return null;
            }

            $crawler = new Crawler($response->body(), $url);

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
            ) ?? now();

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
            Log::debug('The Star article fetch failed', [
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

    protected function getFeedUrls(): array
    {
        return [];
    }

    protected function extractArticleData($item): ?array
    {
        return null;
    }
}

