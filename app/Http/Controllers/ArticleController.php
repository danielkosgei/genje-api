<?php

namespace App\Http\Controllers;

use App\Models\News;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;

class ArticleController extends Controller
{
    /**
     * Show a single article
     */
    public function show($id)
    {
        $article = News::findOrFail($id);
        $related = News::where('category', $article->category)
            ->where('id', '!=', $article->id)
            ->limit(3)
            ->get();

        // Increment views and record history
        try {
            // Increment total views
            $article->increment('views');
            // Increment daily aggregated views
            DB::table('news_views')->updateOrInsert(
                [
                    'news_id' => $article->id,
                    'view_date' => now()->toDateString(),
                ],
                [
                    'views' => DB::raw('views + 1'),
                    'updated_at' => now(),
                    'created_at' => now(),
                ]
            );
        } catch (\Throwable $e) {
            // ignore view counting errors
        }

        if (Auth::check()) {
            try {
                DB::table('reading_history')->insert([
                    'user_id' => Auth::id(),
                    'news_id' => $article->id,
                    'viewed_at' => now(),
                    'created_at' => now(),
                    'updated_at' => now(),
                ]);
            } catch (\Throwable $e) {
                // ignore duplicates
            }
        }

        return view('article', compact('article', 'related'));
    }
}
