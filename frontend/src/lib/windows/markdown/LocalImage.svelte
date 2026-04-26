<script lang="ts">
	import { ReadLocalImage } from '@bindings/nvim-gui/app';

	let { url, altText, class: className = '' }: { url: string; altText: string; class?: string } = $props();

	function isLocalImageUrl(value: string) {
		return (
			value.startsWith('/') ||
			value.startsWith('~') ||
			value.startsWith('file://') ||
			/^[A-Za-z]:[\\/]/.test(value) ||
			value.startsWith('\\\\')
		);
	}

	let localSrc = $state<string | null>(null);
	let isLocal = $derived(isLocalImageUrl(url));

	$effect(() => {
		if (!isLocal || !url) {
			localSrc = null;
			return;
		}
		const requestedUrl = url;
		localSrc = null;
		ReadLocalImage(requestedUrl).then((result) => {
			if (requestedUrl === url) {
				localSrc = result || null;
			}
		});
	});

	let src = $derived(isLocal ? localSrc : url);
</script>

{#if src}
	<img {src} alt={altText} class={className} />
{/if}
