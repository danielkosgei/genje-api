<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\News;
use Illuminate\Http\Request;

class NewsController extends Controller
{
    /**
     * Display a listing of the resource.
     */
    public function index(Request $request)
    {
        $query = News::query()->orderBy('published_at', 'desc');

        // Filter by source
        if ($request->has('source')) {
            $query->where('source', $request->source);
        }

        // Filter by category
        if ($request->has('category')) {
            $query->where('category', $request->category);
        }

        // Search
        if ($request->has('search')) {
            $query->where(function ($q) use ($request) {
                $q->where('title', 'like', '%' . $request->search . '%')
                  ->orWhere('description', 'like', '%' . $request->search . '%')
                  ->orWhere('content', 'like', '%' . $request->search . '%');
            });
        }

        // Pagination
        $perPage = $request->get('per_page', 20);
        $news = $query->paginate($perPage);

        return response()->json($news);
    }

    /**
     * Display the specified resource.
     */
    public function show(string $id)
    {
        $news = News::findOrFail($id);
        return response()->json($news);
    }

    /**
     * Get news by source
     */
    public function bySource($source)
    {
        $news = News::where('source', $source)
            ->orderBy('published_at', 'desc')
            ->paginate(20);
        
        return response()->json($news);
    }

    /**
     * Get news by category
     */
    public function byCategory($category)
    {
        $news = News::where('category', $category)
            ->orderBy('published_at', 'desc')
            ->paginate(20);
        
        return response()->json($news);
    }

    /**
     * Get available sources
     */
    public function sources()
    {
        $sources = News::select('source')
            ->distinct()
            ->pluck('source');
        
        return response()->json($sources);
    }

    /**
     * Search news articles
     */
    public function search(Request $request)
    {
        $search = $request->get('q', $request->get('search'));
        
        if (!$search) {
            return response()->json(['error' => 'Search query is required'], 400);
        }

        $news = News::where(function ($q) use ($search) {
                $q->where('title', 'like', '%' . $search . '%')
                  ->orWhere('description', 'like', '%' . $search . '%')
                  ->orWhere('content', 'like', '%' . $search . '%');
            })
            ->orderBy('published_at', 'desc')
            ->paginate($request->get('per_page', 20));
        
        return response()->json($news);
    }

    /**
     * Get available categories
     */
    public function categories()
    {
        $categories = News::select('category')
            ->whereNotNull('category')
            ->distinct()
            ->pluck('category');
        
        return response()->json($categories);
    }
}
