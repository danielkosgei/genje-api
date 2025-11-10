<?php

namespace App\Jobs;

use App\Models\News;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;

class CacheArticleImageJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    protected int $newsId;

    /**
     * Create a new job instance.
     */
    public function __construct(int $newsId)
    {
        $this->newsId = $newsId;
    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
            $news = News::find($this->newsId);
            if (!$news || !$news->image_url || $news->cached_image_path) {
                return;
            }

            try {
                $response = Http::timeout(20)->get($news->image_url);
                if (!$response->successful()) {
                    return;
                }
                $imageData = $response->body();

                // Try to convert to webp if possible
                $image = @imagecreatefromstring($imageData);
                if ($image !== false && function_exists('imagewebp')) {
                    ob_start();
                    imagepalettetotruecolor($image);
                    imagealphablending($image, true);
                    imagesavealpha($image, true);
                    imagewebp($image, null, 80);
                    $webpData = ob_get_clean();
                    imagedestroy($image);
                    $filename = 'news-images/' . $this->newsId . '-' . md5($news->image_url) . '.webp';
                    Storage::disk('public')->put($filename, $webpData);
                    $news->cached_image_path = $filename;
                    $news->save();
                    return;
                }

                // Fallback: store original
                $ext = 'img';
                if (preg_match('/\\.(jpg|jpeg|png|gif|webp)/i', $news->image_url, $m)) {
                    $ext = strtolower($m[1]);
                }
                $filename = 'news-images/' . $this->newsId . '-' . md5($news->image_url) . '.' . $ext;
                Storage::disk('public')->put($filename, $imageData);
                $news->cached_image_path = $filename;
                $news->save();
            } catch (\Throwable $e) {
                Log::warning('Failed to cache article image', [
                    'news_id' => $this->newsId,
                    'error' => $e->getMessage(),
                ]);
            }
    }
}
