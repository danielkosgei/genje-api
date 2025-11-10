<?php

namespace App\Http\Controllers;

use App\Models\News;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use App\Services\RecommendationService;

class HomeController extends Controller
{
    public function index(Request $request, RecommendationService $recs)
    {
        $query = News::query();

        // Apply search if provided
        if ($request->has('search') && $request->search) {
            $search = $request->search;
            $query->where(function ($q) use ($search) {
                $q->where('title', 'like', '%' . $search . '%')
                  ->orWhere('description', 'like', '%' . $search . '%')
                  ->orWhere('content', 'like', '%' . $search . '%');
            });
        }

        // Apply source filter if provided
        if ($request->has('source') && $request->source) {
            $query->where('source', $request->source);
        }

        // Apply category filter if provided
        if ($request->has('category') && $request->category) {
            $query->where('category', $request->category);
        }

        // Get latest news articles
        $news = $query->orderBy('published_at', 'desc')
            ->paginate(12)
            ->appends($request->query());
        
        // Get distinct sources and categories for filters
        $sources = News::select('source')->distinct()->pluck('source');
        $categories = News::select('category')->whereNotNull('category')->distinct()->pluck('category');
        
        // Favorite IDs for current user (to toggle Save/Unsave)
        $favoriteIds = [];
        $followedSources = [];
        if (Auth::check()) {
            $user = Auth::user();
            $favoriteIds = $user->favoriteNews()
                ->pluck('news.id')
                ->toArray();
            $prefs = $user->preferences ?? [];
            $followedSources = $prefs['followed_sources'] ?? [];
        }

        // Recommended for the user (optional section)
        $recommended = collect();
        if (!($request->has('search') || $request->has('source') || $request->has('category'))) {
            $recommended = $recs->getRecommended(6);
        }

        return view('home', compact('news', 'sources', 'categories', 'favoriteIds', 'recommended', 'followedSources'));
    }
}
