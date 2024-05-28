<script>
	import { SendKey } from '$lib/wailsjs/go/main/App';
	import { EventsOn } from '$lib/wailsjs/runtime/runtime';

	/**@type{string[]}*/
	let lines = [];

	function send_key(e) {
		SendKey(e.key, e.ctrlKey, e.shiftKey, e.altKey);
	}

	EventsOn('buf-lines-changed', (updatedLines) => {
		lines = updatedLines;
		console.log('buflineschanged', updatedLines);
	});
</script>

<div class="h-screen w-screen">
	<h1>Buffer</h1>
	<input on:keydown|preventDefault={send_key} type="text" />
	{#each lines as line}
		{#if !line}
			<div class="whitespace-pre">{'\u00A0'}</div>
		{:else}
			<div class="whitespace-pre">{@html line}</div>
		{/if}
	{/each}
</div>
