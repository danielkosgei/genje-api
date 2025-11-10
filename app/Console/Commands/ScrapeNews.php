<?php

namespace App\Console\Commands;

use App\Services\NewsScraperManager;
use Illuminate\Console\Command;

class ScrapeNews extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'news:scrape {--source= : Scrape a specific news source}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Scrape news articles from Kenyan news sources';

    /**
     * Execute the console command.
     */
    public function handle(NewsScraperManager $scraperManager): int
    {
        $this->info('Starting news scraping...');

        $source = $this->option('source');

        if ($source) {
            $this->info("Scraping from: {$source}");
            $count = $scraperManager->scrapeSource($source);
            
            if ($count === null) {
                $this->error("Source '{$source}' not found.");
                return Command::FAILURE;
            }

            $this->info("Scraped {$count} articles from {$source}");
            return Command::SUCCESS;
        }

        $this->info('Scraping from all sources...');
        $results = $scraperManager->scrapeAll();

        $this->newLine();
        $this->info('Scraping Results:');
        $this->newLine();

        $totalCount = 0;
        foreach ($results as $source => $result) {
            if ($result['success']) {
                $this->line("✓ {$source}: {$result['count']} articles");
                $totalCount += $result['count'];
            } else {
                $this->error("✗ {$source}: {$result['error']}");
            }
        }

        $this->newLine();
        $this->info("Total articles scraped: {$totalCount}");

        return Command::SUCCESS;
    }
}
