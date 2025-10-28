<?php

namespace App\Http\Controllers;

use App\Models\News;
use Illuminate\Http\Request;

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

        return view('article', compact('article', 'related'));
    }
}
