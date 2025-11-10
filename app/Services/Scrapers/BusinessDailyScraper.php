<?php

namespace App\Services\Scrapers;

class BusinessDailyScraper extends BaseScraper
{
    public function getSourceName(): string
    {
        return 'Business Daily';
    }

    protected function getFeedUrls(): array
    {
        return [
            'https://www.businessdailyafrica.com/rss',
            'https://www.businessdailyafrica.com/rss/news',
            'https://www.businessdailyafrica.com/rss/business',
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

        // Business Daily is primarily business news
        $category = $this->extractCategory($link);

        return [
            'title' => $this->cleanHtml($title),
            'description' => $this->cleanHtml($description) ?: null,
            'content' => $this->cleanHtml($description) ?: null,
            'source' => $this->getSourceName(),
            'category' => $category ?: 'business',
            'url' => $link,
            'image_url' => $this->extractImageUrl($item),
            'author' => $this->cleanHtml($author) ?: null,
            'published_at' => $this->rssService->parseDate($pubDate) ?? now(),
        ];
    }

    private function extractCategory(string $url): ?string
    {
        $urlLower = strtolower($url);
        
        if (strpos($urlLower, '/markets/') !== false) {
            return 'business';
        }
        if (strpos($urlLower, '/companies/') !== false) {
            return 'business';
        }
        if (strpos($urlLower, '/economy/') !== false) {
            return 'business';
        }
        if (strpos($urlLower, '/technology/') !== false) {
            return 'technology';
        }

        return 'business';
    }
}

