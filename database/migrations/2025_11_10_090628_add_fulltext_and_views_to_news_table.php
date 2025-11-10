<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::table('news', function (Blueprint $table) {
            if (!Schema::hasColumn('news', 'views')) {
                $table->unsignedBigInteger('views')->default(0)->after('published_at');
            }
        });

        $driver = DB::getDriverName();
        if ($driver === 'mysql') {
            DB::statement('ALTER TABLE news ADD FULLTEXT news_fulltext (title, description, content)');
        } elseif ($driver === 'pgsql') {
            DB::statement("CREATE INDEX IF NOT EXISTS news_fulltext_idx ON news USING GIN (to_tsvector('english', coalesce(title,'') || ' ' || coalesce(description,'') || ' ' || coalesce(content,'')))");
        }
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        $driver = DB::getDriverName();
        if ($driver === 'mysql') {
            DB::statement('ALTER TABLE news DROP INDEX news_fulltext');
        } elseif ($driver === 'pgsql') {
            DB::statement('DROP INDEX IF EXISTS news_fulltext_idx');
        }

        Schema::table('news', function (Blueprint $table) {
            if (Schema::hasColumn('news', 'views')) {
                $table->dropColumn('views');
            }
        });
    }
};
