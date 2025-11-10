<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;

class UserPreferenceController extends Controller
{
    public function followSource(Request $request, string $source)
    {
        $user = Auth::user();
        $prefs = $user->preferences ?? [];
        $followed = $prefs['followed_sources'] ?? [];
        if (!in_array($source, $followed, true)) {
            $followed[] = $source;
        }
        $prefs['followed_sources'] = array_values($followed);
        $user->preferences = $prefs;
        $user->save();

        if ($request->wantsJson()) {
            return response()->json(['status' => 'followed', 'source' => $source]);
        }
        return back()->with('status', "Following {$source}");
    }

    public function unfollowSource(Request $request, string $source)
    {
        $user = Auth::user();
        $prefs = $user->preferences ?? [];
        $followed = $prefs['followed_sources'] ?? [];
        $prefs['followed_sources'] = array_values(array_filter($followed, fn ($s) => $s !== $source));
        $user->preferences = $prefs;
        $user->save();

        if ($request->wantsJson()) {
            return response()->json(['status' => 'unfollowed', 'source' => $source]);
        }
        return back()->with('status', "Unfollowed {$source}");
    }
}
