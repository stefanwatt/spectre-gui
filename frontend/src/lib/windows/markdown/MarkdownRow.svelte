<script lang="ts">
	const headingHls = [587, 589, 480, 479, 330, 477];
	let {
		row,
		decode
	}: {
		row: App.NvimRow;
		decode?: (input: string) => string;
	} = $props();
</script>

{#each row.tokens as token}
	{#if token?.text}
		{#if headingHls.includes(token.highlight)}
			<span class="cell hl-{token.highlight} inline-block h-full {token?.classes}">
				{decode ? decode(token.text.replace(/^#+\s*/, '')) : token.text.replace(/^#+\s*/, '')}
			</span>
		{:else}
			<span class="cell hl-{token.highlight} inline-block h-full {token?.classes}">
				{decode ? decode(token.text) : token.text}
			</span>
		{/if}
	{:else}
		<span class="cell inline-block h-full">⠀</span>
	{/if}
{:else}
	<span class="cell inline-block h-full">⠀</span>
{/each}

<style>
	.cell {
		padding-top: 3px;
		padding-bottom: 3px;
	}

	/*h1*/
	.hl-587 {
		font-size: 50.29px; /* 32 / 14 * 22 */
		line-height: 62.86px; /* 40 / 14 * 22 */
	}

	/*h2*/
	.hl-589 {
		font-size: 37.71px; /* 24 / 14 * 22 */
		line-height: 47.14px; /* 30 / 14 * 22 */
	}

	.hl-480 {
		font-size: 31.43px; /* 20 / 14 * 22 */
		line-height: 39.29px; /* 25 / 14 * 22 */
	}

	.hl-479 {
		font-size: 25.14px; /* 16 / 14 * 22 */
		line-height: 31.43px; /* 20 / 14 * 22 */
	}

	.hl-330 {
		font-family: 22px; /* same as base */
		line-height: 27.5px; /* 17.5 / 14 * 22 */
	}

	.hl-477 {
		font-size: 21.37px; /* 13.6 / 14 * 22 */
		line-height: 26.71px; /* 17 / 14 * 22 */
	}
</style>
