<?php

namespace App\Http\Controllers;

use App\Models\News;
use Illuminate\Http\Request;

class HomeController extends Controller
{
    public function index(Request $request)
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
        
        return view('home', compact('news', 'sources', 'categories'));
    }
}
