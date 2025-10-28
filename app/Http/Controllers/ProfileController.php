<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;

class ProfileController extends Controller
{
    /**
     * Show the user's profile
     */
    public function show()
    {
        return view('profile');
    }
}
