<!DOCTYPE html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}">
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
                        <title>{{ config('app.name', 'Genje') }} - Saved</title>
        
        <link rel="icon" href="/favicon.ico" sizes="any">
        <link rel="icon" href="/favicon.svg" type="image/svg+xml">
        <link rel="apple-touch-icon" href="/apple-touch-icon.png">

        <link rel="preconnect" href="https://fonts.bunny.net">
        <link href="https://fonts.bunny.net/css?family=instrument-sans:400,500,600" rel="stylesheet" />

        @if (file_exists(public_path('build/manifest.json')) || file_exists(public_path('hot')))
            @vite(['resources/css/app.css', 'resources/js/app.js'])
        @else
            <style>
                /*! Tailwind minimal fallback (same block as home uses) */
            </style>
        @endif
    </head>
    <body class="bg-[#FDFDFC] dark:bg-[#0a0a0a] text-[#1b1b18] dark:text-[#EDEDEC] min-h-screen">
        <nav class="border-b border-[#e3e3e0] dark:border-[#3E3E3A] bg-white dark:bg-[#161615] sticky top-0 z-50">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 py-3 sm:py-4">
                <div class="flex items-center justify-between">
                    <div class="flex items-center gap-4 sm:gap-8">
                        <a href="{{ route('home') }}" class="text-lg sm:text-xl font-semibold">
                            {{ config('app.name', 'Genje') }}
                        </a>
                    </div>
                    <div class="flex items-center gap-2 sm:gap-4">
                        @auth
                        <a href="{{ route('favorites.index') }}" class="text-sm hover:opacity-80">Saved</a>
                        <a href="{{ route('logout') }}" class="text-sm hover:opacity-80 text-red-600 dark:text-red-400">Logout</a>
                        @else
                        <a href="{{ route('auth.google') }}" class="px-3 sm:px-4 py-2 text-xs sm:text-sm font-medium bg-[#1b1b18] dark:bg:white text:white dark:text-[#1b1b18] border border-[#1b1b18] dark:border-white">
                            Login
                        </a>
                        @endauth
                    </div>
                </div>
            </div>
        </nav>

        <main class="max-w-7xl mx-auto px-4 sm:px-6 py-8 sm:py-12">
            <div class="mb-8">
                <h1 class="text-2xl sm:text-3xl font-semibold">My Saved Articles</h1>
                <p class="text-[#706f6c] dark:text-[#A1A09A] mt-2">Articles you've saved to read later.</p>
            </div>

            @if($news->count() > 0)
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
                @foreach($news as $article)
                <div class="bg-white dark:bg-[#161615] border border-[#e3e3e0] dark:border-[#3E3E3A]">
                    @php
                        $img = $article->cached_image_path ? Storage::url($article->cached_image_path) : $article->image_url;
                    @endphp
                    @if($img)
                    <div class="w-full h-40 sm:h-48 bg-gray-200 dark:bg-[#2a2a2a] bg-cover bg-center" style="background-image: url('{{ $img }}');"></div>
                    @endif
                    <div class="p-4 sm:p-6">
                        <div class="flex items-center justify-between gap-2 mb-2">
                            <div class="flex items-center gap-2">
                                <span class="text-xs font-medium text-[#706f6c] dark:text-[#A1A09A]">{{ $article->source }}</span>
                                @if($article->category)
                                <span class="text-xs px-2 py-1 bg-[#f5f5f5] dark:bg-[#2a2a2a] text-[#706f6c] dark:text-[#A1A09A] capitalize">
                                    {{ $article->category }}
                                </span>
                                @endif
                            </div>
                            <form method="POST" action="{{ route('favorites.destroy', $article->id) }}">
                                @csrf
                                @method('DELETE')
                                <button class="text-xs px-2 py-1 border border-[#e3e3e0] dark:border-[#3E3E3A] hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a]">Unsave</button>
                            </form>
                        </div>
                        <a href="{{ route('article', $article->id) }}" class="block">
                            <h2 class="text-lg sm:text-xl font-semibold mb-2 line-clamp-2">{{ $article->title }}</h2>
                            <p class="text-sm text-[#706f6c] dark:text-[#A1A09A] mb-4 line-clamp-3">{{ $article->description }}</p>
                        </a>
                    </div>
                </div>
                @endforeach
            </div>

            <div class="mt-8">
                {{ $news->links() }}
            </div>
            @else
            <div class="text-center py-16">
                <p class="text-[#706f6c] dark:text-[#A1A09A]">You haven't saved any articles yet.</p>
            </div>
            @endif
        </main>
    </body>
</html>

