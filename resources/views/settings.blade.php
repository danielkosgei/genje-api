<!DOCTYPE html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}">
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <title>Settings - {{ config('app.name', 'Genje') }}</title>
        
        <link rel="icon" href="/favicon.ico" sizes="any">
        <link rel="icon" href="/favicon.svg" type="image/svg+xml">
        <link rel="apple-touch-icon" href="/apple-touch-icon.png">

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
        <main class="max-w-4xl mx-auto px-4 sm:px-6 py-8 sm:py-12">
            <div class="bg-white dark:bg-[#161615] border border-[#e3e3e0] dark:border-[#3E3E3A] p-6 sm:p-8">
                <h1 class="text-3xl font-semibold mb-6">Settings</h1>
                
                <div class="space-y-8">
                    <!-- App Preferences -->
                    <div>
                        <h2 class="text-xl font-semibold mb-4">Preferences</h2>
                        <div class="space-y-4">
                            <div class="flex items-center justify-between border-b border-[#e3e3e0] dark:border-[#3E3E3A] pb-4">
                                <div>
                                    <p class="font-medium text-[#1b1b18] dark:text-[#EDEDEC]">Language</p>
                                    <p class="text-sm text-[#706f6c] dark:text-[#A1A09A]">English</p>
                                </div>
                                <button class="px-4 py-2 text-sm border border-[#e3e3e0] dark:border-[#3E3E3A] hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a] transition-colors">Change</button>
                            </div>
                            
                            <div class="flex items-center justify-between border-b border-[#e3e3e0] dark:border-[#3E3E3A] pb-4">
                                <div>
                                    <p class="font-medium text-[#1b1b18] dark:text-[#EDEDEC]">Dark Mode</p>
                                    <p class="text-sm text-[#706f6c] dark:text-[#A1A09A]">Follow system</p>
                                </div>
                                <button class="px-4 py-2 text-sm border border-[#e3e3e0] dark:border-[#3E3E3A] hover:bg-[#f5f5f5] dark:hover:bg-[#2a2a2a] transition-colors">Change</button>
                            </div>
                        </div>
                    </div>

                    <!-- Account Settings -->
                    <div>
                        <h2 class="text-xl font-semibold mb-4">Account</h2>
                        <div class="space-y-4">
                            <div>
                                <p class="font-medium text-[#1b1b18] dark:text-[#EDEDEC] mb-2">Login Provider</p>
                                <p class="text-sm text-[#706f6c] dark:text-[#A1A09A]">Connected via {{ ucfirst(auth()->user()->provider) }}</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
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

