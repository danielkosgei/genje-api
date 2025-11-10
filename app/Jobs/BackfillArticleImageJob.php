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

class BackfillArticleImageJob implements ShouldQueue
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
        if (!$news || $news->image_url) {
            return;
        }

        try {
            $finalUrl = $this->resolveArticleUrl($news->url);
            if (!$finalUrl) {
                return;
            }

            $resp = Http::timeout(20)->withHeaders([
                'User-Agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
                'Accept' => 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
            ])->get($finalUrl);
            if (!$resp->successful()) {
                return;
            }
            $html = $resp->body();

            $publisherHost = parse_url($finalUrl, PHP_URL_HOST) ?: '';

            $candidates = [];
            // og:image
            if (preg_match('/<meta\s+property=["\']og:image["\']\s+content=["\']([^"\']+)["\']/i', $html, $m)) {
                $candidates[] = trim($m[1]);
            }
            // og:image:secure_url
            if (preg_match('/<meta\s+property=["\']og:image:secure_url["\']\s+content=["\']([^"\']+)["\']/i', $html, $m)) {
                $candidates[] = trim($m[1]);
            }
            // twitter:image
            if (preg_match('/<meta\s+name=["\']twitter:image["\']\s+content=["\']([^"\']+)["\']/i', $html, $m)) {
                $candidates[] = trim($m[1]);
            }

            // Fallback: first <img> in article body
            if (empty($candidates)) {
                if (preg_match('/<article[\s\S]*?<img[^>]+src=["\']([^"\']+)["\']/i', $html, $m)) {
                    $candidates[] = trim($m[1]);
                } elseif (preg_match('/<img[^>]+src=["\']([^"\']+)["\'][^>]*class=["\'][^"\']*(hero|featured|main)[^"\']*["\']/i', $html, $m)) {
                    $candidates[] = trim($m[1]);
                } elseif (preg_match('/<img[^>]+src=["\']([^"\']+)["\']/i', $html, $m)) {
                    $candidates[] = trim($m[1]);
                }
            }

            $image = $this->selectPublisherImage($candidates, $publisherHost);
            if (!$image) {
                return;
            }

            $news->image_url = $image;
            $news->save();
            \App\Jobs\CacheArticleImageJob::dispatch($news->id);
        } catch (\Throwable $e) {
            Log::warning('BackfillArticleImageJob failed', [
                'news_id' => $this->newsId,
                'error' => $e->getMessage(),
            ]);
        }
    }

    private function resolveArticleUrl(string $url): ?string
    {
        try {
            $resp = Http::timeout(20)->withHeaders([
                'User-Agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
            ])->get($url);
            if (!$resp->successful()) {
                return $url;
            }
            $final = (string) $resp->effectiveUri();
            // If still on news.google.com, try canonical/og:url
            if (strpos($final, 'news.google.') !== false) {
                $html = $resp->body();
                if (preg_match('/<link\s+rel=["\']canonical["\']\s+href=["\']([^"\']+)["\']/i', $html, $m)) {
                    return $m[1];
                }
                if (preg_match('/<meta\s+property=["\']og:url["\']\s+content=["\']([^"\']+)["\']/i', $html, $m)) {
                    return $m[1];
                }
            }
            return $final;
        } catch (\Throwable $e) {
            return $url;
        }
    }

    private function selectPublisherImage(array $candidates, string $publisherHost): ?string
    {
        foreach ($candidates as $url) {
            $host = parse_url($url, PHP_URL_HOST) ?: '';
            $lower = strtolower($url);
            // Exclude Google/gstatic/cache or sprite/logo-like images
            if (strpos($host, 'google') !== false || strpos($host, 'gstatic') !== false) {
                continue;
            }
            if (preg_match('/(sprite|logo|icon|placeholder|default)/i', $lower)) {
                continue;
            }
            // Prefer same-domain or CDN subdomains
            if ($publisherHost && (strpos($host, $publisherHost) !== false)) {
                return $url;
            }
        }
        // Otherwise return first non-google candidate
        foreach ($candidates as $url) {
            $host = parse_url($url, PHP_URL_HOST) ?: '';
            if (strpos($host, 'google') === false && strpos($host, 'gstatic') === false) {
                return $url;
            }
        }
        return null;
    }
}
