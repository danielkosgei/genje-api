<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Relations\BelongsToMany;

class News extends Model
{
    use HasFactory;

    protected $fillable = [
        'title',
        'description',
        'content',
        'source',
        'category',
        'url',
        'fingerprint',
        'image_url',
        'cached_image_path',
        'author',
        'quality_score',
        'published_at',
    ];

    protected $casts = [
        'published_at' => 'datetime',
    ];

    public function favoritedBy(): BelongsToMany
    {
        return $this->belongsToMany(User::class, 'favorites', 'news_id', 'user_id')
            ->withTimestamps();
    }
}
