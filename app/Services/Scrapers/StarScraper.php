<?php

namespace App\Services\Scrapers;

class StarScraper extends BaseScraper
{
    public function getSourceName(): string
    {
        return 'The Star';
    }

    protected function getFeedUrls(): array
    {
        return [
            'https://www.the-star.co.ke/rss',
            'https://www.the-star.co.ke/rss/news',
            'https://www.the-star.co.ke/rss/business',
            'https://www.the-star.co.ke/rss/sports',
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

        $category = $this->extractCategory($link);

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

    private function extractCategory(string $url): ?string
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

