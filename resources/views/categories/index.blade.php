<!DOCTYPE html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}">
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
                        <title>{{ config('app.name', 'Genje') }} - Categories</title>
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
                    <a href="{{ route('favorites.index') }}" class="text-sm hover:opacity-80 hidden sm:inline">Saved</a>
                </div>
            </div>
        </nav>

        <main class="max-w-7xl mx-auto px-4 sm:px-6 py-8 sm:py-12">
            <h1 class="text-2xl sm:text-3xl font-semibold mb-6">Browse by Category</h1>

            @if($categories->count() > 0)
            <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3 sm:gap-4">
                @foreach($categories as $cat)
                <a href="{{ route('categories.show', $cat) }}" class="px-3 py-2 border border-[#e3e3e0] dark:border-[#3E3E3A] bg-white dark:bg-[#161615] hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a] transition-colors flex items-center justify-between">
                    <span class="capitalize">{{ $cat }}</span>
                    <span class="text-xs text-[#706f6c] dark:text-[#A1A09A]">{{ $counts[$cat] ?? 0 }}</span>
                </a>
                @endforeach
            </div>
            @else
            <div class="text-center py-16">
                <p class="text-[#706f6c] dark:text-[#A1A09A]">No categories available.</p>
            </div>
            @endif
        </main>
    </body>
</html>

