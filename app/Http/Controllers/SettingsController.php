<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;

class SettingsController extends Controller
{
    /**
     * Show the settings page
     */
    public function index()
    {
        $user = Auth::user();
        $preferences = $user->preferences ?? [];
        return view('settings', compact('preferences'));
    }

    /**
     * Save preferences
     */
    public function update(Request $request)
    {
        $validated = $request->validate([
            'email_notifications' => 'nullable|boolean',
            'language' => 'nullable|string|max:10',
        ]);

        $user = Auth::user();
        $prefs = $user->preferences ?? [];
        $prefs['email_notifications'] = (bool) ($validated['email_notifications'] ?? false);
        $prefs['language'] = $validated['language'] ?? ($prefs['language'] ?? 'en');

        $user->preferences = $prefs;
        $user->save();

        if ($request->wantsJson()) {
            return response()->json(['status' => 'saved', 'preferences' => $prefs]);
        }
        return back()->with('status', 'Preferences saved.');
    }
}
