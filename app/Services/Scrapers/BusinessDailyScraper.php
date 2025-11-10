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
            'https://www.businessdailyafrica.com/bd/rss.xml',
        ];
    }

    protected function extractArticleData($item): ?array
    {
        $title = $this->rssService->getText($item, './/title');
        $link = $this->extractLinkFromItem($item);
        $descriptionRaw = $this->getDecodedDescription($item);
        $pubDate = $this->rssService->getText($item, './/pubDate');
        $author = $this->extractAuthorFromItem($item);

        if ($title === '' || !$link) {
            return null;
        }

        $description = $descriptionRaw ? $this->cleanHtml($descriptionRaw) : null;
        $category = $this->extractCategoryFromItem($item) ?? 'business';

        return [
            'title' => $this->cleanHtml($title),
            'description' => $description ?: null,
            'content' => $description ?: null,
            'source' => $this->getSourceName(),
            'category' => $category ?: 'business',
            'url' => $link,
            'image_url' => $this->extractImageUrl($item),
            'author' => $author,
            'published_at' => $this->rssService->parseDate($pubDate) ?? now(),
        ];
    }
}

