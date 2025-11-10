<?php

namespace App\Http\Controllers;

use App\Models\News;
use Illuminate\Http\Request;

class CategoryController extends Controller
{
    public function index()
    {
        $categories = News::select('category')
            ->whereNotNull('category')
            ->distinct()
            ->orderBy('category')
            ->pluck('category');

        // Counts per category
        $counts = News::selectRaw('category, COUNT(*) as total')
            ->whereNotNull('category')
            ->groupBy('category')
            ->pluck('total', 'category');

        return view('categories.index', compact('categories', 'counts'));
    }

    public function show(Request $request, string $category)
    {
        $news = News::where('category', $category)
            ->when($request->get('search'), function ($q) use ($request) {
                $search = $request->get('search');
                $q->where(function ($qq) use ($search) {
                    $qq->where('title', 'like', '%' . $search . '%')
                       ->orWhere('description', 'like', '%' . $search . '%')
                       ->orWhere('content', 'like', '%' . $search . '%');
                });
            })
            ->orderByDesc('published_at')
            ->paginate(12)
            ->appends($request->query());

        return view('categories.show', compact('news', 'category'));
    }
}
