<?php

namespace App\Services;

use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use SimpleXMLElement;

class RssFeedService
{
    /**
     * Fetch and parse RSS feed
     */
    public function fetchFeed(string $url): ?SimpleXMLElement
    {
        try {
            $response = Http::timeout(30)
                ->withHeaders([
                    'User-Agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
                ])
                ->get($url);

            if (!$response->successful()) {
                Log::warning("Failed to fetch RSS feed: {$url}", [
                    'status' => $response->status(),
                ]);
                return null;
            }

            $xml = @simplexml_load_string($response->body());
            
            if ($xml === false) {
                Log::warning("Failed to parse RSS feed XML: {$url}");
                return null;
            }

            return $xml;
        } catch (\Exception $e) {
            Log::error("Error fetching RSS feed: {$url}", [
                'error' => $e->getMessage(),
            ]);
            return null;
        }
    }

    /**
     * Extract items from RSS feed
     */
    public function extractItems(SimpleXMLElement $xml): array
    {
        $items = [];

        // Handle RSS 2.0 format
        if (isset($xml->channel->item)) {
            foreach ($xml->channel->item as $item) {
                $items[] = $item;
            }
        }
        // Handle Atom format
        elseif (isset($xml->entry)) {
            foreach ($xml->entry as $entry) {
                $items[] = $entry;
            }
        }

        return $items;
    }

    /**
     * Get text content from XML element
     */
    public function getText(SimpleXMLElement $element, string $path, string $default = ''): string
    {
        // Try direct property access first (for RSS 2.0)
        $parts = explode('/', $path);
        $lastPart = end($parts);
        $lastPart = str_replace('.//', '', $lastPart);
        
        if (isset($element->{$lastPart})) {
            return trim((string) $element->{$lastPart});
        }

        // Try xpath
        $namespaces = $element->getNamespaces(true);
        $result = $element->xpath($path);

        if ($result && count($result) > 0) {
            return trim((string) $result[0]);
        }

        // Try with namespaces
        foreach ($namespaces as $prefix => $namespace) {
            $result = $element->xpath("{$prefix}:{$lastPart}");
            if ($result && count($result) > 0) {
                return trim((string) $result[0]);
            }
        }

        // For Atom feeds, try different paths
        if ($lastPart === 'link') {
            // Atom feeds use link[@href]
            $result = $element->xpath('.//link[@href]');
            if ($result && count($result) > 0) {
                $attrs = $result[0]->attributes();
                if (isset($attrs['href'])) {
                    return trim((string) $attrs['href']);
                }
            }
        }

        return $default;
    }

    /**
     * Get attribute value from XML element
     */
    public function getAttribute(SimpleXMLElement $element, string $path, string $attribute, string $default = ''): string
    {
        $result = $element->xpath($path);

        if ($result && count($result) > 0) {
            $attrs = $result[0]->attributes();
            if (isset($attrs[$attribute])) {
                return trim((string) $attrs[$attribute]);
            }
        }

        return $default;
    }

    /**
     * Parse date string to Carbon instance
     */
    public function parseDate(string $dateString): ?\Carbon\Carbon
    {
        try {
            return \Carbon\Carbon::parse($dateString);
        } catch (\Exception $e) {
            Log::warning("Failed to parse date: {$dateString}", [
                'error' => $e->getMessage(),
            ]);
            return null;
        }
    }
}

