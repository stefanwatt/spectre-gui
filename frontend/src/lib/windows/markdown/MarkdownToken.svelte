<script lang="ts">
	const imageExtensions = new Set(['png', 'jpg']);
	const highlights = {
		headings: new Set([587, 589, 480, 479, 330, 477]),
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
	{#if highlights.headings.has(token.highlight)}
		<span class="cell hl-{token.highlight} inline-block h-full {token?.classes}">
			{text.replace(/^(1)\s*/, '')}
		</span>
	{:else if token.highlight == highlights.task.completed || token.highlight == highlights.task.uncompleted}
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

	/*h1*/
	.hl-1945,
	.hl-1944 {
		font-size: 50.29px; /* 32 / 14 * 22 */
		line-height: 62.86px; /* 40 / 14 * 22 */
	}

	/*h2*/
	.hl-1964,
	.hl-1963 {
		font-size: 37.71px; /* 24 / 14 * 22 */
		line-height: 47.14px; /* 30 / 14 * 22 */
	}

	.hl-1966,
	.hl-1967 {
		font-size: 31.43px; /* 20 / 14 * 22 */
		line-height: 39.29px; /* 25 / 14 * 22 */
	}

	.hl-1969,
	.hl-1970 {
		font-size: 25.14px; /* 16 / 14 * 22 */
		line-height: 31.43px; /* 20 / 14 * 22 */
	}

	.hl-1972,
	.hl-1973 {
		font-family: 22px; /* same as base */
		line-height: 27.5px; /* 17.5 / 14 * 22 */
	}

	.hl-1976,
	.hl-1975 {
		font-size: 21.37px; /* 13.6 / 14 * 22 */
		line-height: 26.71px; /* 17 / 14 * 22 */
	}
</style>
