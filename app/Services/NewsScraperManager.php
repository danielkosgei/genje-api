<?php

namespace App\Services;

use App\Services\Scrapers\BusinessDailyScraper;
use App\Services\Scrapers\CitizenTvScraper;
use App\Services\Scrapers\DailyNationScraper;
use App\Services\Scrapers\StandardScraper;
use App\Services\Scrapers\StarScraper;
use Illuminate\Support\Facades\Log;

class NewsScraperManager
{
    protected RssFeedService $rssService;

    public function __construct(RssFeedService $rssService)
    {
        $this->rssService = $rssService;
    }

    /**
     * Get all available scrapers
     */
    public function getScrapers(): array
    {
        return [
            new DailyNationScraper($this->rssService),
            new StandardScraper($this->rssService),
            new CitizenTvScraper($this->rssService),
            new BusinessDailyScraper($this->rssService),
            new StarScraper($this->rssService),
        ];
    }

    /**
     * Run all scrapers
     */
    public function scrapeAll(): array
    {
        $results = [];
        $scrapers = $this->getScrapers();

        foreach ($scrapers as $scraper) {
            try {
                $count = $scraper->scrape();
                $results[$scraper->getSourceName()] = [
                    'success' => true,
                    'count' => $count,
                ];
            } catch (\Exception $e) {
                Log::error("Scraper failed: {$scraper->getSourceName()}", [
                    'error' => $e->getMessage(),
                    'trace' => $e->getTraceAsString(),
                ]);
                $results[$scraper->getSourceName()] = [
                    'success' => false,
                    'error' => $e->getMessage(),
                ];
            }
        }

        return $results;
    }

    /**
     * Run a specific scraper by source name
     */
    public function scrapeSource(string $sourceName): ?int
    {
        $scrapers = $this->getScrapers();

        foreach ($scrapers as $scraper) {
            if ($scraper->getSourceName() === $sourceName) {
                try {
                    return $scraper->scrape();
                } catch (\Exception $e) {
                    Log::error("Scraper failed: {$sourceName}", [
                        'error' => $e->getMessage(),
                    ]);
                    throw $e;
                }
            }
        }

        return null;
    }
}

