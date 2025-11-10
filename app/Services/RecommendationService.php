<?php

namespace App\Services;

use App\Models\News;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;

class RecommendationService
{
    /**
     * Get recommended articles for the current user based on recent reading history.
     * Uses simple source/category frequency scoring over the last 14 days.
     */
    public function getRecommended(int $limit = 12)
    {
        $user = Auth::user();
        if (!$user) {
            return collect();
        }

        $prefs = $user->preferences ?? [];
        $followedSources = $prefs['followed_sources'] ?? [];

        $since = now()->subDays(14);

        // Build frequency maps
        $history = DB::table('reading_history')
            ->join('news', 'reading_history.news_id', '=', 'news.id')
            ->where('reading_history.user_id', $user->id)
            ->where('reading_history.viewed_at', '>=', $since)
            ->select('news.source', 'news.category')
            ->get();

        $sourceScores = [];
        $categoryScores = [];
        foreach ($history as $row) {
            if ($row->source) {
                $sourceScores[$row->source] = ($sourceScores[$row->source] ?? 0) + 2; // weight sources more
            }
            if ($row->category) {
                $categoryScores[$row->category] = ($categoryScores[$row->category] ?? 0) + 1;
            }
        }

        // Boost followed sources heavily
        foreach ($followedSources as $src) {
            $sourceScores[$src] = ($sourceScores[$src] ?? 0) + 5;
        }

        if (empty($sourceScores) && empty($categoryScores)) {
            return collect();
        }

        // Build query: newer first, then apply score ordering
        $query = News::query()
            ->orderByDesc('published_at')
            ->limit($limit * 3); // fetch more to sort by score

        $candidates = $query->get();

        // Score candidates
        $scored = $candidates->map(function (News $n) use ($sourceScores, $categoryScores) {
            $score = 0;
            if ($n->source && isset($sourceScores[$n->source])) {
                $score += $sourceScores[$n->source];
            }
            if ($n->category && isset($categoryScores[$n->category])) {
                $score += $categoryScores[$n->category];
            }
            // small freshness boost for last 48h
            if ($n->published_at && now()->diffInHours($n->published_at) <= 48) {
                $score += 1;
            }
            return [$n, $score];
        })->filter(fn ($pair) => $pair[1] > 0)
          ->sortByDesc(fn ($pair) => $pair[1])
          ->take($limit)
          ->map(fn ($pair) => $pair[0])
          ->values();

        return $scored;
    }
}
