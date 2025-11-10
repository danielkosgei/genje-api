<?php

use App\Http\Controllers\AuthController;
use App\Http\Controllers\HomeController;
use App\Http\Controllers\ProfileController;
use App\Http\Controllers\SettingsController;
use App\Http\Controllers\ArticleController;
use App\Http\Controllers\FavoritesController;
use App\Http\Controllers\CategoryController;
use App\Http\Controllers\UserPreferenceController;
use Illuminate\Support\Facades\Route;

Route::get('/', [HomeController::class, 'index'])->name('home');

// Authentication routes
Route::get('/auth/google', [AuthController::class, 'redirect'])->name('auth.google');
Route::get('/auth/google/callback', [AuthController::class, 'callback'])->name('auth.google.callback');
Route::get('/logout', [AuthController::class, 'logout'])->name('logout');

// Protected routes
Route::middleware('auth')->group(function () {
    Route::get('/profile', [ProfileController::class, 'show'])->name('profile');
    Route::get('/settings', [SettingsController::class, 'index'])->name('settings');
    Route::post('/settings', [SettingsController::class, 'update'])->name('settings.update');

    // Favorites
    Route::get('/favorites', [FavoritesController::class, 'index'])->name('favorites.index');
    Route::post('/favorites/{news}', [FavoritesController::class, 'store'])->name('favorites.store');
    Route::delete('/favorites/{news}', [FavoritesController::class, 'destroy'])->name('favorites.destroy');

    // Source preferences
    Route::post('/preferences/source/{source}', [UserPreferenceController::class, 'followSource'])->name('preferences.source.follow');
    Route::delete('/preferences/source/{source}', [UserPreferenceController::class, 'unfollowSource'])->name('preferences.source.unfollow');
});

// Article routes
Route::get('/article/{id}', [ArticleController::class, 'show'])->name('article');

// Category routes
Route::get('/categories', [CategoryController::class, 'index'])->name('categories.index');
Route::get('/categories/{category}', [CategoryController::class, 'show'])->name('categories.show');
