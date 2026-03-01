<script lang="ts">
	const imageExtensions = new Set(['png', 'jpg']);
	const highlights = {
		task: { completed: 168, uncompleted: 167 },
		url: 170
	};
	let {
		token,
		decode
	}: {
		token: App.NvimToken;
		decode?: (input: string) => string;
	} = $props();
	let text = $derived(decode ? decode(token.text) : token.text);
	let isImageUrl = $derived(
		(() => {
			const matches = text.match(/\.([0-9a-z]+)(?:[\?#]|$)/i);
			if (!matches) return false;
			return imageExtensions.has(matches[1]);
		})()
	);
</script>

{#if text}
	{#if token.highlight == highlights.task.completed || token.highlight == highlights.task.uncompleted}
		<div class="flex items-center p-1">
			<input
				type="checkbox"
				checked={token.highlight == highlights.task.completed}
				class="checkbox"
			/>
			<span class="cell hl-{token.highlight} inline-block h-full {token?.classes}">
				{text.slice(3)}
			</span>
		</div>
	{:else if token.highlight == highlights.url}
		{#if isImageUrl}
			<img src={text} alt="image" />
		{:else}
			<a href={text}>{text}</a>
		{/if}
	{:else}
		<span class="cell hl-{token.highlight} inline-block h-full {token?.classes}">
			{text}
		</span>
	{/if}
{:else}
	<span class="cell inline-block h-full">⠀</span>
{/if}

<style>
	.cell {
		padding-top: 3px;
		padding-bottom: 3px;
	}
</style>
