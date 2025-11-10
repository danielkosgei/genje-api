<?php

namespace App\Http\Controllers;

use App\Models\News;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;

class FavoritesController extends Controller
{
    public function index(Request $request)
    {
        $user = Auth::user();
        $favorites = $user->favoriteNews()
            ->orderByDesc('news.published_at')
            ->paginate(12);

        return view('favorites.index', [
            'news' => $favorites,
        ]);
    }

    public function store(Request $request, $newsId)
    {
        $request->validate([
            'news_id' => 'nullable', // supports both path param and form
        ]);

        $user = Auth::user();
        $news = News::findOrFail($newsId);
        $user->favoriteNews()->syncWithoutDetaching([$news->id]);

        if ($request->wantsJson()) {
            return response()->json(['status' => 'saved']);
        }

        return back()->with('status', 'Article saved.');
    }

    public function destroy(Request $request, $newsId)
    {
        $user = Auth::user();
        $news = News::findOrFail($newsId);
        $user->favoriteNews()->detach($news->id);

        if ($request->wantsJson()) {
            return response()->json(['status' => 'removed']);
        }

        return back()->with('status', 'Removed from saved.');
    }
}
