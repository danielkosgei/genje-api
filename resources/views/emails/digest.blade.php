@php
$articles = $articles ?? [];
@endphp
<x-mail::message>
# Your Daily News Digest

Hi {{ $name }},

Here are some stories we think you'll enjoy:

@foreach($articles as $article)
- **{{ $article->title }}** ({{ $article->source }})
  - {{ $article->description }}
  - [Read more]({{ url(route('article', $article->id)) }})
@endforeach

@if(empty($articles))
We couldn't find new articles today based on your interests. Check back soon!
@endif

Thanks,<br>
{{ config('app.name', 'Genje') }}
</x-mail::message>
