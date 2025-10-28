<!DOCTYPE html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}">
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <title>{{ config('app.name', 'Laravel') }}</title>
        
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
            </style>
        @endif
    </head>
    <body class="bg-[#FDFDFC] dark:bg-[#0a0a0a] text-[#1b1b18] dark:text-[#EDEDEC] min-h-screen">
        <!-- Navigation -->
        <nav class="border-b border-[#e3e3e0] dark:border-[#3E3E3A] bg-white dark:bg-[#161615]">
            <div class="max-w-7xl mx-auto px-6 py-4">
                <div class="flex items-center justify-between">
                    <div class="flex items-center gap-8">
                        <a href="{{ route('home') }}" class="text-xl font-semibold">
                            {{ config('app.name', 'Laravel') }}
                        </a>
                    </div>
                    
                    <div class="flex items-center gap-4">
                        @auth
                            <span class="text-sm text-[#706f6c] dark:text-[#A1A09A]">
                                Welcome, <span class="font-medium text-[#1b1b18] dark:text-[#EDEDEC]">{{ auth()->user()->name }}</span>
                            </span>
                            <a href="{{ route('logout') }}" class="px-4 py-2 text-sm font-medium border border-black dark:border-white hover:bg-black hover:text-white dark:hover:bg-white dark:hover:text-[#1b1b18] transition-colors">
                                Logout
                            </a>
                        @else
                            <a href="{{ route('auth.google') }}" class="px-4 py-2 text-sm font-medium bg-[#1b1b18] dark:bg-white text-white dark:text-[#1b1b18] hover:opacity-90 transition-opacity">
                                Login with Google
                            </a>
                        @endauth
                    </div>
                </div>
            </div>
        </nav>

        <!-- Main Content -->
        <main class="max-w-7xl mx-auto px-6 py-16">
            <div class="text-center mb-12">
                <h1 class="text-5xl font-semibold mb-4">
                    @auth
                        Welcome back!
                    @else
                        Get Started
                    @endauth
                </h1>
                <p class="text-lg text-[#706f6c] dark:text-[#A1A09A]">
                    @auth
                        You're successfully logged in.
                    @else
                        Sign in with your Google account to continue.
                    @endauth
                </p>
            </div>

            @auth
                <div class="mt-12 p-8 bg-white dark:bg-[#161615] border border-[#e3e3e0] dark:border-[#3E3E3A] rounded-lg">
                    <h2 class="text-2xl font-semibold mb-4">Your Account</h2>
                    <div class="space-y-2">
                        <div class="flex justify-between">
                            <span class="text-[#706f6c] dark:text-[#A1A09A]">Name:</span>
                            <span class="font-medium">{{ auth()->user()->name }}</span>
                        </div>
                        <div class="flex justify-between">
                            <span class="text-[#706f6c] dark:text-[#A1A09A]">Email:</span>
                            <span class="font-medium">{{ auth()->user()->email }}</span>
                        </div>
                        <div class="flex justify-between">
                            <span class="text-[#706f6c] dark:text-[#A1A09A]">Provider:</span>
                            <span class="font-medium">{{ ucfirst(auth()->user()->provider) }}</span>
                        </div>
                    </div>
                </div>
            @else
                <div class="mt-12 p-8 bg-white dark:bg-[#161615] border border-[#e3e3e0] dark:border-[#3E3E3A] rounded-lg text-center">
                    <p class="text-[#706f6c] dark:text-[#A1A09A] mb-6">
                        Click the button above to sign in with Google.
                    </p>
                </div>
            @endauth
        </main>

        <!-- Footer -->
        <footer class="border-t border-[#e3e3e0] dark:border-[#3E3E3A] mt-16">
            <div class="max-w-7xl mx-auto px-6 py-6">
                <p class="text-sm text-center text-[#706f6c] dark:text-[#A1A09A]">
                    © {{ date('Y') }} {{ config('app.name', 'Laravel') }}. All rights reserved.
                </p>
            </div>
        </footer>
    </body>
</html>

