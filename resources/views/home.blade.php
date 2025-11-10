<!DOCTYPE html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}">
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
                        <title>{{ config('app.name', 'Genje') }}</title>
        
        <link rel="icon" href="/favicon.ico" sizes="any">
        <link rel="icon" href="/favicon.svg" type="image/svg+xml">
        <link rel="apple-touch-icon" href="/apple-touch-icon.png">

        <!-- Fonts -->
        <link rel="preconnect" href="https://fonts.bunny.net">
        <link href="https://fonts.bunny.net/css?family=instrument-sans:400,500,600" rel="stylesheet" />

        <!-- Styles / Scripts -->
        @if (file_exists(public_path('build/manifest.json')) || file_exists(public_path('hot')))
            @vite(['resources/css/app.css', 'resources/js/app.js'])
        @else
            <style>
                /*! tailwindcss v4.0.7 | MIT License | https://tailwindcss.com */@layer theme{:root,:host{--font-sans:'Instrument Sans',ui-sans-serif,system-ui,sans-serif,"Apple Color Emoji","Segoe UI Emoji","Segoe UI Symbol","Noto Color Emoji";--font-serif:ui-serif,Georgia,Cambria,"Times New Roman",Times,serif;--font-mono:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono","Courier New",monospace;--color-red-50:oklch(.971 .013 17.38);--color-red-100:oklch(.936 .032 17.717);--color-red-200:oklch(.885 .062 18.334);--color-red-300:oklch(.808 .114 19.571);--color-red-400:oklch(.704 .191 22.216);--color-red-500:oklch(.637 .237 25.331);--color-red-600:oklch(.577 .245 27.325);--color-red-700:oklch(.505 .213 27.518);--color-red-800:oklch(.444 .177 26.899);--color-red-900:oklch(.396 .141 25.723);--color-red-950:oklch(.258 .092 26.042);--color-slate-50:oklch(.984 .003 247.858);--color-slate-100:oklch(.968 .007 247.896);--color-slate-200:oklch(.929 .013 255.508);--color-slate-300:oklch(.869 .022 252.894);--color-slate-400:oklch(.704 .04 256.788);--color-slate-500:oklch(.554 .046 257.417);--color-slate-600:oklch(.446 .043 257.281);--color-slate-700:oklch(.372 .044 257.287);--color-slate-800:oklch(.279 .041 260.031);--color-slate-900:oklch(.208 .042 265.755);--color-slate-950:oklch(.129 .042 264.695);--color-black:#000;--color-white:#fff;--spacing:.25rem;--text-base:1rem;--text-base--line-height: 1.5 ;--font-weight-medium:500;--font-weight-semibold:600;--shadow-sm:0 1px 3px 0 #0000001a,0 1px 2px -1px #0000001a;--default-transition-duration:.15s;--default-transition-timing-function:cubic-bezier(.4,0,.2,1)}@layer base{*,:after,:before,::backdrop{box-sizing:border-box;border:0 solid;margin:0;padding:0}html,:host{line-height:1.5;font-family:var(--font-sans)}body{line-height:inherit}}
            [x-cloak] { display: none !important; }
            </style>
        @endif
    </head>
    <body class="bg-[#FDFDFC] dark:bg-[#0a0a0a] text-[#1b1b18] dark:text-[#EDEDEC] min-h-screen">
        <!-- Navigation -->
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
                            <!-- Profile Dropdown -->
                            <div class="relative" x-data="{ open: false }">
                                <button @click="open = !open" class="flex items-center gap-2 hover:opacity-80 transition-opacity focus:outline-none">
                                    @if(auth()->user()->avatar)
                                        <img src="{{ auth()->user()->avatar }}" alt="{{ auth()->user()->name }}" class="w-8 h-8 border border-[#e3e3e0] dark:border-[#3E3E3A] object-cover">
                                    @else
                                        <div class="w-8 h-8 bg-[#1b1b18] dark:bg-white text-white dark:text-[#1b1b18] flex items-center justify-center font-medium text-xs border border-[#e3e3e0] dark:border-[#3E3E3A]">
                                            {{ auth()->user()->initials() }}
                                        </div>
                                    @endif
                                </button>
                                
                                <!-- Dropdown Menu -->
                                <div x-show="open" @click.away="open = false" x-cloak class="absolute right-0 mt-2 w-56 bg-white dark:bg-[#161615] border border-[#e3e3e0] dark:border-[#3E3E3A] shadow-lg z-50" x-transition>
                                    <div class="p-4 border-b border-[#e3e3e0] dark:border-[#3E3E3A]">
                                        <p class="text-sm font-medium text-[#1b1b18] dark:text-[#EDEDEC]">{{ auth()->user()->name }}</p>
                                        <p class="text-xs text-[#706f6c] dark:text-[#A1A09A] truncate">{{ auth()->user()->email }}</p>
                                    </div>
                                    <div class="p-2 space-y-1">
                                        <a href="{{ route('favorites.index') }}" class="block px-4 py-2 text-sm hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a] transition-colors text-[#1b1b18] dark:text-[#EDEDEC]">
                                            Saved
                                        </a>
                                        <a href="{{ route('profile') }}" class="block px-4 py-2 text-sm hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a] transition-colors text-[#1b1b18] dark:text-[#EDEDEC]">
                                            Profile
                                        </a>
                                        <a href="{{ route('settings') }}" class="block px-4 py-2 text-sm hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a] transition-colors text-[#1b1b18] dark:text-[#EDEDEC]">
                                            Settings
                                        </a>
                                        <div class="border-t border-[#e3e3e0] dark:border-[#3E3E3A] my-1"></div>
                                        <a href="{{ route('logout') }}" class="block px-4 py-2 text-sm hover:bg-red-50 dark:hover:bg-red-950/20 transition-colors text-red-600 dark:text-red-400">
                                            Logout
                                        </a>
                                    </div>
                                </div>
                            </div>
                        @else
                            <a href="{{ route('auth.google') }}" class="px-3 sm:px-4 py-2 text-xs sm:text-sm font-medium bg-[#1b1b18] dark:bg-white text-white dark:text-[#1b1b18] hover:opacity-90 transition-opacity border border-[#1b1b18] dark:border-white">
                                <span class="hidden sm:inline">Login with Google</span>
                                <span class="sm:hidden">Login</span>
                            </a>
                        @endauth
                    </div>
                </div>
            </div>
        </nav>

        <!-- Main Content -->
        <main class="max-w-7xl mx-auto px-4 sm:px-6 py-8 sm:py-12">
            <!-- Hero Section -->
            <div class="text-center mb-8 sm:mb-12">
                <h1 class="text-3xl sm:text-4xl md:text-5xl font-semibold mb-3 sm:mb-4">
                    Latest News
                </h1>
                <p class="text-base sm:text-lg text-[#706f6c] dark:text-[#A1A09A] mb-6 sm:mb-8">
                    Stay informed with the latest news
                </p>

                <!-- Search Bar -->
                <form method="GET" action="{{ route('home') }}" class="max-w-2xl mx-auto">
                    <div class="flex gap-2 flex-col sm:flex-row">
                        <input 
                            type="text" 
                            name="search" 
                            value="{{ request('search') }}"
                            placeholder="Search news..." 
                            class="flex-1 px-4 sm:px-6 py-2 sm:py-3 text-sm sm:text-base border border-[#e3e3e0] dark:border-[#3E3E3A] bg-white dark:bg-[#161615] focus:outline-none focus:ring-2 focus:ring-black dark:focus:ring-white"
                        >
                        <button 
                            type="submit" 
                            class="px-6 sm:px-8 py-2 sm:py-3 text-sm sm:text-base bg-[#1b1b18] dark:bg-white text-white dark:text-[#1b1b18] font-medium hover:opacity-90 transition-opacity whitespace-nowrap"
                        >
                            Search
                        </button>
                        @if(request('search'))
                        <a href="{{ route('home') }}" class="px-4 sm:px-6 py-2 sm:py-3 text-sm sm:text-base border border-[#e3e3e0] dark:border-[#3E3E3A] hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a] transition-colors whitespace-nowrap">
                            Clear
                        </a>
                        @endif
                    </div>
                </form>
            </div>

            @if(request('search'))
            <div class="mb-6 sm:mb-8 text-center px-4">
                <p class="text-sm sm:text-base text-[#706f6c] dark:text-[#A1A09A]">
                    Search results for "<strong class="text-[#1b1b18] dark:text-[#EDEDEC]">{{ request('search') }}</strong>" 
                    <span class="ml-2">({{ $news->total() }} {{ Str::plural('result', $news->total()) }})</span>
                </p>
            </div>
            @endif

            <!-- Filters -->
            @if(isset($sources) && $sources->count() > 0)
            <div class="mb-6 sm:mb-8 flex gap-2 sm:gap-3 justify-center flex-wrap px-2">
                <a href="{{ route('home', request('search') ? ['search' => request('search')] : []) }}" class="px-3 sm:px-4 py-2 text-xs sm:text-sm border border-[#e3e3e0] dark:border-[#3E3E3A] hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a] transition-colors {{ !request('source') ? 'bg-[#1b1b18] dark:bg-white text-white dark:text-[#1b1b18] border-transparent' : '' }}">
                    All
                </a>
                @foreach($sources as $source)
                <a href="{{ route('home', array_filter(['source' => $source, 'search' => request('search')])) }}" class="px-3 sm:px-4 py-2 text-xs sm:text-sm border border-[#e3e3e0] dark:border-[#3E3E3A] hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a] transition-colors {{ request('source') === $source ? 'bg-[#1b1b18] dark:bg-white text-white dark:text-[#1b1b18] border-transparent' : '' }}">
                    {{ $source }}
                </a>
                @endforeach
            </div>
            @endif

            <!-- News Grid -->
            @if($news->count() > 0)
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6 mb-8 sm:mb-12">
                @foreach($news as $article)
                <div class="bg-white dark:bg-[#161615] border border-[#e3e3e0] dark:border-[#3E3E3A] hover:shadow-lg transition-shadow">
                    @if($article->image_url)
                    <div class="w-full h-40 sm:h-48 bg-gray-200 dark:bg-[#2a2a2a] bg-cover bg-center" style="background-image: url('{{ $article->image_url }}');"></div>
                    @endif
                    <div class="p-4 sm:p-6">
                        <div class="flex items-center gap-2 mb-2 flex-wrap justify-between">
                            <span class="text-xs font-medium text-[#706f6c] dark:text-[#A1A09A]">{{ $article->source }}</span>
                            @if($article->category)
                            <span class="text-xs px-2 py-1 bg-[#f5f5f5] dark:bg-[#2a2a2a] text-[#706f6c] dark:text-[#A1A09A] capitalize">
                                {{ $article->category }}
                            </span>
                            @endif
                            @auth
                            <div class="ml-auto">
                                @php $isSaved = isset($favoriteIds) && in_array($article->id, $favoriteIds, true); @endphp
                                @if($isSaved)
                                <form method="POST" action="{{ route('favorites.destroy', $article->id) }}">
                                    @csrf
                                    @method('DELETE')
                                    <button type="submit" class="text-xs px-2 py-1 border border-[#e3e3e0] dark:border-[#3E3E3A] hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a]">
                                        Unsave
                                    </button>
                                </form>
                                @else
                                <form method="POST" action="{{ route('favorites.store', $article->id) }}">
                                    @csrf
                                    <button type="submit" class="text-xs px-2 py-1 border border-[#e3e3e0] dark:border-[#3E3E3A] hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a]">
                                        Save
                                    </button>
                                </form>
                                @endif
                            </div>
                            @endauth
                        </div>
                        <a href="{{ route('article', $article->id) }}" class="block">
                            <h2 class="text-lg sm:text-xl font-semibold mb-2 line-clamp-2">{{ $article->title }}</h2>
                            <p class="text-sm text-[#706f6c] dark:text-[#A1A09A] mb-4 line-clamp-3">{{ $article->description }}</p>
                        </a>
                        <div class="flex items-center justify-between flex-wrap gap-2">
                            <span class="text-xs text-[#706f6c] dark:text-[#A1A09A]">
                                {{ $article->published_at->diffForHumans() }}
                            </span>
                            <span class="text-xs font-medium text-[#1b1b18] dark:text-[#EDEDEC]">
                                Read more →
                            </span>
                        </div>
                    </div>
                </div>
                @endforeach
            </div>

            <!-- Pagination -->
            @if($news->hasPages())
            <div class="flex justify-center gap-2 mt-8 flex-wrap">
                @if ($news->onFirstPage())
                <span class="px-3 sm:px-4 py-2 text-xs sm:text-sm border border-[#e3e3e0] dark:border-[#3E3E3A] text-[#706f6c] dark:text-[#A1A09A] cursor-not-allowed">
                    Previous
                </span>
                @else
                <a href="{{ $news->previousPageUrl() }}" class="px-3 sm:px-4 py-2 text-xs sm:text-sm border border-[#e3e3e0] dark:border-[#3E3E3A] hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a] transition-colors">
                    Previous
                </a>
                @endif

                <span class="px-3 sm:px-4 py-2 text-xs sm:text-sm border border-[#e3e3e0] dark:border-[#3E3E3A]">
                    Page {{ $news->currentPage() }} of {{ $news->lastPage() }}
                </span>

                @if ($news->hasMorePages())
                <a href="{{ $news->nextPageUrl() }}" class="px-3 sm:px-4 py-2 text-xs sm:text-sm border border-[#e3e3e0] dark:border-[#3E3E3A] hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a] transition-colors">
                    Next
                </a>
                @else
                <span class="px-3 sm:px-4 py-2 text-xs sm:text-sm border border-[#e3e3e0] dark:border-[#3E3E3A] text-[#706f6c] dark:text-[#A1A09A] cursor-not-allowed">
                    Next
                </span>
                @endif
            </div>
            @endif
            @else
            <div class="text-center py-16">
                <p class="text-[#706f6c] dark:text-[#A1A09A]">No news articles found.</p>
            </div>
            @endif
        </main>

        <!-- Footer -->
        <footer class="border-t border-[#e3e3e0] dark:border-[#3E3E3A] mt-12 sm:mt-16">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 py-4 sm:py-6">
                <p class="text-xs sm:text-sm text-center text-[#706f6c] dark:text-[#A1A09A]">
                    © {{ date('Y') }} {{ config('app.name', 'Genje') }}. All rights reserved.
                </p>
            </div>
        </footer>
    </body>
</html>
