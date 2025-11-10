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
            'https://news.google.com/rss/search?q=site:nation.africa&hl=en-KE&gl=KE&ceid=KE:en',
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
        $category = $this->extractCategoryFromItem($item);

        return [
            'title' => $this->cleanHtml($title),
            'description' => $description ?: null,
            'content' => $description ?: null,
            'source' => $this->getSourceName(),
            'category' => $category,
            'url' => $link,
            'image_url' => $this->extractImageUrl($item),
            'author' => $author,
            'published_at' => $this->rssService->parseDate($pubDate) ?? now(),
        ];
    }
}

