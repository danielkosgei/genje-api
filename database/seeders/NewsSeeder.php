<?php

namespace Database\Seeders;

use App\Models\News;
use Illuminate\Database\Console\Seeds\WithoutModelEvents;
use Illuminate\Database\Seeder;

class NewsSeeder extends Seeder
{
    /**
     * Run the database seeds.
     */
    public function run(): void
    {
        $newsItems = [
            [
                'title' => 'Kenya to Host Major Climate Summit',
                'description' => 'Nairobi will host an international climate summit bringing together leaders from across Africa.',
                'content' => 'Nairobi has been chosen as the venue for the upcoming African Climate Summit. The event will focus on renewable energy and sustainable development across the continent.',
                'source' => 'Daily Nation',
                'category' => 'environment',
                'url' => 'https://nation.africa/news/kenya-climate-summit',
                'image_url' => 'https://images.pexels.com/photos/33148790/pexels-photo-33148790.jpeg',
                'author' => 'John Kamau',
                'published_at' => now()->subHours(2),
            ],
            [
                'title' => 'Kenya\'s Economy Shows Strong Growth',
                'description' => 'Kenya\'s GDP growth exceeded expectations, driven by agriculture and technology sectors.',
                'content' => 'The latest economic data shows Kenya\'s economy is growing faster than projected. The agricultural sector and tech industry are leading the charge.',
                'source' => 'The Standard',
                'category' => 'business',
                'url' => 'https://standardmedia.co.ke/business/kenya-economy-growth',
                'image_url' => 'https://images.pexels.com/photos/34205396/pexels-photo-34205396.jpeg',
                'author' => 'Mary Wanjiku',
                'published_at' => now()->subHours(5),
            ],
            [
                'title' => 'Government Unveils New Housing Plan',
                'description' => 'Affordable housing initiative promises thousands of new units for Kenyan families.',
                'content' => 'The government has announced an ambitious affordable housing program targeting urban and rural areas. This initiative aims to address the housing deficit in Kenya.',
                'source' => 'NTV',
                'category' => 'politics',
                'url' => 'https://ntv.co.ke/politics/housing-plan',
                'image_url' => 'https://images.pexels.com/photos/18911015/pexels-photo-18911015.jpeg',
                'author' => 'Peter Njenga',
                'published_at' => now()->subHours(8),
            ],
            [
                'title' => 'Kenya Wins African Cup of Nations Qualifier',
                'description' => 'Harambee Stars triumph in crucial match against Ethiopia.',
                'content' => 'Kenya\'s national football team secured a vital victory in their quest for African Cup of Nations qualification. The team showed great determination and skill.',
                'source' => 'Citizen TV',
                'category' => 'sports',
                'url' => 'https://citizentv.co.ke/sports/kenya-afcon',
                'image_url' => 'https://images.pexels.com/photos/34427557/pexels-photo-34427557.jpeg',
                'author' => 'James Ochieng',
                'published_at' => now()->subHours(12),
            ],
            [
                'title' => 'New Tech Hub Opens in Nairobi',
                'description' => 'State-of-the-art innovation center to support local startups and entrepreneurs.',
                'content' => 'Nairobi\'s tech ecosystem received a major boost with the opening of a new innovation hub in Kilimani. The facility will support local startups and foster technological innovation.',
                'source' => 'BBC Africa',
                'category' => 'technology',
                'url' => 'https://bbc.com/africa/kenya-tech-hub',
                'image_url' => 'https://images.pexels.com/photos/32948694/pexels-photo-32948694.jpeg',
                'author' => 'Sarah Mutua',
                'published_at' => now()->subHours(1),
            ],
        ];

        foreach ($newsItems as $news) {
            News::create($news);
        }
    }
}
