<?php

namespace App\Console\Commands;

use App\Jobs\BackfillArticleImageJob;
use App\Models\News;
use Illuminate\Console\Command;

class BackfillMissingImages extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'images:backfill {--limit=200 : Max number of articles to scan}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Backfill missing article images by fetching og:image from source pages';

    /**
     * Execute the console command.
     */
    public function handle()
    {
        $limit = (int) $this->option('limit');
        $this->info("Scanning for articles missing images (limit {$limit})...");

        $articles = News::whereNull('image_url')
            ->orderByDesc('published_at')
            ->limit($limit)
            ->get();

        foreach ($articles as $article) {
            BackfillArticleImageJob::dispatch($article->id);
        }

        $this->info("Dispatched backfill jobs for {$articles->count()} articles.");
        return Command::SUCCESS;
    }
}
