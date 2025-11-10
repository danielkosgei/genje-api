<?php

namespace App\Services\Scrapers;

class DailyNationScraper extends BaseScraper
{
    public function getSourceName(): string
    {
        return 'Daily Nation';
    }

    protected function getFeedUrls(): array
    {
        return [
            'https://nation.africa/rss',
            'https://nation.africa/rss/news',
            'https://nation.africa/rss/business',
            'https://nation.africa/rss/sports',
        ];
    }

    protected function extractArticleData($item): ?array
    {
        $title = $this->rssService->getText($item, './/title');
        $link = $this->rssService->getText($item, './/link');
        $description = $this->rssService->getText($item, './/description');
        $pubDate = $this->rssService->getText($item, './/pubDate');
        $author = $this->rssService->getText($item, './/author') ?: $this->rssService->getText($item, './/dc:creator');

        if (!$title || !$link) {
            return null;
        }

        // Extract category from link or description
        $category = $this->extractCategory($link, $description);

        return [
            'title' => $this->cleanHtml($title),
            'description' => $this->cleanHtml($description) ?: null,
            'content' => $this->cleanHtml($description) ?: null,
            'source' => $this->getSourceName(),
            'category' => $category,
            'url' => $link,
            'image_url' => $this->extractImageUrl($item),
            'author' => $this->cleanHtml($author) ?: null,
            'published_at' => $this->rssService->parseDate($pubDate) ?? now(),
        ];
    }

    private function extractCategory(string $url, string $description): ?string
    {
        $urlLower = strtolower($url);
        
        if (strpos($urlLower, '/sports/') !== false) {
            return 'sports';
        }
        if (strpos($urlLower, '/business/') !== false) {
            return 'business';
        }
        if (strpos($urlLower, '/politics/') !== false) {
            return 'politics';
        }
        if (strpos($urlLower, '/technology/') !== false) {
            return 'technology';
        }
        if (strpos($urlLower, '/health/') !== false) {
            return 'health';
        }

        return null;
    }
}

