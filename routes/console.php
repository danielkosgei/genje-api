<?php

use Illuminate\Foundation\Inspiring;
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Schedule;

Artisan::command('inspire', function () {
    $this->comment(Inspiring::quote());
})->purpose('Display an inspiring quote');

// Schedule news scraping every hour
Schedule::command('news:scrape')
    ->hourly()
    ->withoutOverlapping()
    ->runInBackground();

Schedule::command('digest:send')
    ->dailyAt('07:00')
    ->withoutOverlapping();
