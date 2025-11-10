<?php

namespace App\Console\Commands;

use App\Mail\DailyDigestMail;
use App\Models\User;
use App\Services\RecommendationService;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Mail;

class SendDailyDigest extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'digest:send';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Send daily news digests to opted-in users';

    /**
     * Execute the console command.
     */
    public function handle(RecommendationService $recs)
    {
        $this->info('Preparing daily digests...');

        $users = User::query()
            ->whereNotNull('email')
            ->get()
            ->filter(function (User $user) {
                $prefs = $user->preferences ?? [];
                return ($prefs['email_notifications'] ?? false) === true;
            });

        foreach ($users as $user) {
            $articles = $recs->getRecommended(5);
            if ($articles->isEmpty()) {
                continue;
            }

            Mail::to($user->email)->send(new DailyDigestMail($user->name, $articles));
            $this->line("Sent digest to {$user->email}");
        }

        $this->info('Daily digests complete.');
        return Command::SUCCESS;
    }
}
