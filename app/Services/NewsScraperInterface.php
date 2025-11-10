<?php

namespace App\Services;

interface NewsScraperInterface
{
    /**
     * Get the source name
     */
    public function getSourceName(): string;

    /**
     * Scrape news articles from the source
     */
    public function scrape(): int;
}

