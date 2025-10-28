<?php

use App\Http\Controllers\Api\NewsController;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;

Route::get('/user', function (Request $request) {
    return $request->user();
})->middleware('auth:sanctum');

// News API routes
Route::prefix('news')->group(function () {
    Route::get('/', [NewsController::class, 'index']);
    Route::get('/search', [NewsController::class, 'search']);
    Route::get('/meta/sources', [NewsController::class, 'sources']);
    Route::get('/meta/categories', [NewsController::class, 'categories']);
    Route::get('/source/{source}', [NewsController::class, 'bySource']);
    Route::get('/category/{category}', [NewsController::class, 'byCategory']);
    Route::get('/{id}', [NewsController::class, 'show']);
});
