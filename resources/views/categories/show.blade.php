<!DOCTYPE html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}">
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
                        <title>{{ config('app.name', 'Genje') }} - {{ ucfirst($category) }}</title>
        @if (file_exists(public_path('build/manifest.json')) || file_exists(public_path('hot')))
            @vite(['resources/css/app.css', 'resources/js/app.js'])
        @endif
    </head>
    <body class="bg-[#FDFDFC] dark:bg-[#0a0a0a] text-[#1b1b18] dark:text-[#EDEDEC] min-h-screen">
        <nav class="border-b border-[#e3e3e0] dark:border-[#3E3E3A] bg-white dark:bg-[#161615] sticky top-0 z-50">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 py-3 sm:py-4">
                <div class="flex items-center justify-between">
                    <a href="{{ route('home') }}" class="text-lg sm:text-xl font-semibold">
                        {{ config('app.name', 'Genje') }}
                    </a>
                    <a href="{{ route('categories.index') }}" class="text-sm hover:opacity-80 hidden sm:inline">Categories</a>
                </div>
            </div>
        </nav>

        <main class="max-w-7xl mx-auto px-4 sm:px-6 py-8 sm:py-12">
            <div class="mb-6 sm:mb-8 flex items-center justify-between">
                <h1 class="text-2xl sm:text-3xl font-semibold capitalize">{{ $category }}</h1>
                <form method="GET" action="{{ route('categories.show', $category) }}" class="flex gap-2">
                    <input type="text" name="search" value="{{ request('search') }}" placeholder="Search in {{ $category }}..." class="px-4 py-2 text-sm border border-[#e3e3e0] dark:border-[#3E3E3A] bg-white dark:bg-[#161615]">
                    <button type="submit" class="px-4 py-2 text-sm bg-[#1b1b18] dark:bg-white text-white dark:text-[#1b1b18]">Search</button>
                </form>
            </div>

            @if($news->count() > 0)
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
                @foreach($news as $article)
                <a href="{{ route('article', $article->id) }}" class="bg-white dark:bg-[#161615] border border-[#e3e3e0] dark:border-[#3E3E3A] hover:shadow-lg transition-shadow">
                    @if($article->image_url)
                    <div class="w-full h-40 sm:h-48 bg-gray-200 dark:bg-[#2a2a2a] bg-cover bg-center" style="background-image: url('{{ $article->image_url }}');"></div>
                    @endif
                    <div class="p-4 sm:p-6">
                        <div class="flex items-center gap-2 mb-2 flex-wrap">
                            <span class="text-xs font-medium text-[#706f6c] dark:text-[#A1A09A]">{{ $article->source }}</span>
                            @if($article->category)
                            <span class="text-xs px-2 py-1 bg-[#f5f5f5] dark:bg-[#2a2a2a] text-[#706f6c] dark:text-[#A1A09A] capitalize">
                                {{ $article->category }}
                            </span>
                            @endif
                        </div>
                        <h2 class="text-lg sm:text-xl font-semibold mb-2 line-clamp-2">{{ $article->title }}</h2>
                        <p class="text-sm text-[#706f6c] dark:text-[#A1A09A] mb-4 line-clamp-3">{{ $article->description }}</p>
                        <div class="flex items-center justify-between flex-wrap gap-2">
                            <span class="text-xs text-[#706f6c] dark:text-[#A1A09A]">
                                {{ $article->published_at->diffForHumans() }}
                            </span>
                            <span class="text-xs font-medium text-[#1b1b18] dark:text-[#EDEDEC]">
                                Read more →
                            </span>
                        </div>
                    </div>
                </a>
                @endforeach
            </div>

            <div class="mt-8">
                {{ $news->links() }}
            </div>
            @else
            <div class="text-center py-16">
                <p class="text-[#706f6c] dark:text-[#A1A09A]">No articles found in this category.</p>
            </div>
            @endif
        </main>
    </body>
</html>

