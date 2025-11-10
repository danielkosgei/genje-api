<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::table('news', function (Blueprint $table) {
            if (!Schema::hasColumn('news', 'fingerprint')) {
                $table->string('fingerprint')->nullable()->after('url');
                $table->unique('fingerprint', 'news_fingerprint_unique');
            }
            if (!Schema::hasColumn('news', 'quality_score')) {
                $table->unsignedSmallInteger('quality_score')->default(0)->after('author');
            }
            if (!Schema::hasColumn('news', 'cached_image_path')) {
                $table->string('cached_image_path')->nullable()->after('image_url');
            }
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::table('news', function (Blueprint $table) {
                if (Schema::hasColumn('news', 'fingerprint')) {
                    $table->dropUnique('news_fingerprint_unique');
                    $table->dropColumn('fingerprint');
                }
                if (Schema::hasColumn('news', 'quality_score')) {
                    $table->dropColumn('quality_score');
                }
                if (Schema::hasColumn('news', 'cached_image_path')) {
                    $table->dropColumn('cached_image_path');
                }
        });
    }
};
