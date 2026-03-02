<script lang="ts">
	import { ReadLocalImage } from '$lib/wailsjs/go/main/App';

	let { url, altText, class: className = '' }: { url: string; altText: string; class?: string } = $props();

	const LOCAL_IMAGE_PREFIX = '/local-image/';

	let dataUrl = $state<string | null>(null);
	let isLocal = $derived(url.startsWith(LOCAL_IMAGE_PREFIX));

	$effect(() => {
		if (!isLocal) {
			dataUrl = null;
			return;
		}
		const encoded = url.slice(LOCAL_IMAGE_PREFIX.length);
		dataUrl = null;
		ReadLocalImage(encoded).then((result) => {
			dataUrl = result || null;
		});
	});

	let src = $derived(isLocal ? dataUrl : url);
</script>

{#if src}
	<img {src} alt={altText} class={className} />
{/if}
