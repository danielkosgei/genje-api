<?php

namespace App\Services;

use App\Models\News;
use Illuminate\Support\Collection;
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
        [$sourceScores, $categoryScores] = $this->buildPreferenceScores($user);

        $query = News::query()
            ->orderByDesc('published_at')
            ->limit($limit * 3); // fetch more to sort by score

        $candidates = $query->get();

        $ranked = $this->rankArticles($candidates, $sourceScores, $categoryScores);

        return $ranked->take($limit);
    }

    /**
     * Rank a collection of articles, placing highly recommended items first.
     * Falls back to recency for items with equal recommendation scores.
     */
    public function rankArticles(Collection $articles, array $sourceScores = [], array $categoryScores = []): Collection
    {
        $now = now();

        // If no explicit scores provided, derive from current user (if any)
        if (empty($sourceScores) && empty($categoryScores)) {
            $user = Auth::user();
            [$sourceScores, $categoryScores] = $this->buildPreferenceScores($user);
        }

        $scored = $articles->map(function (News $article) use ($sourceScores, $categoryScores, $now) {
            $score = $this->scoreArticle($article, $sourceScores, $categoryScores);

            // small freshness boost for last 48h
            if ($article->published_at && $now->diffInHours($article->published_at) <= 48) {
                $score += 1;
            }

            return ['article' => $article, 'score' => $score];
        });

        $sorted = $scored->sort(function (array $a, array $b) {
            if ($a['score'] === $b['score']) {
                $aTime = optional($a['article']->published_at)->timestamp ?? 0;
                $bTime = optional($b['article']->published_at)->timestamp ?? 0;
                return $bTime <=> $aTime; // newer first
            }

            return $b['score'] <=> $a['score']; // higher score first
        });

        return $sorted->map(fn (array $row) => $row['article'])->values();
    }

    protected function buildPreferenceScores($user): array
    {
        if (!$user) {
            return [[], []];
        }

        $prefs = $user->preferences ?? [];
        $followedSources = $prefs['followed_sources'] ?? [];

        $since = now()->subDays(14);

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
                $sourceScores[$row->source] = ($sourceScores[$row->source] ?? 0) + 2;
            }
            if ($row->category) {
                $categoryScores[$row->category] = ($categoryScores[$row->category] ?? 0) + 1;
            }
        }

        foreach ($followedSources as $src) {
            $sourceScores[$src] = ($sourceScores[$src] ?? 0) + 5;
        }

        return [$sourceScores, $categoryScores];
    }

    protected function scoreArticle(News $article, array $sourceScores, array $categoryScores): int
    {
        $score = 0;

        if ($article->source && isset($sourceScores[$article->source])) {
            $score += $sourceScores[$article->source];
        }

        if ($article->category && isset($categoryScores[$article->category])) {
            $score += $categoryScores[$article->category];
        }

        return $score;
    }
}
