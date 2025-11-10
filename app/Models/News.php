<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Factories\HasFactory;

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
}
