<?php

namespace App\Services\Scrapers;

class StandardScraper extends BaseScraper
{
    public function getSourceName(): string
    {
        return 'The Standard';
    }

    protected function getFeedUrls(): array
    {
        return [
            'https://www.standardmedia.co.ke/rss/headlines.php',
            'https://www.standardmedia.co.ke/rss/kenya.php',
            'https://www.standardmedia.co.ke/rss/politics.php',
            'https://www.standardmedia.co.ke/rss/business.php',
            'https://www.standardmedia.co.ke/rss/sports.php',
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

