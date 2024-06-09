<script>
	/**@type{{
    children:any
    items: App.BufLine[];
    containerHeight:number
    itemHeight:number
    scrollTop:number
  }} */
	let { children, items, containerHeight, itemHeight, scrollTop } = $props();

	let spacerHeight = $derived(Math.max(containerHeight, items.length * itemHeight));
	let numItems = $derived(Math.ceil(containerHeight / itemHeight) + 1);
	let startIndex = $derived(Math.floor(scrollTop / itemHeight));
	let endIndex = $derived(startIndex + numItems);
	let numOverlap = $derived(startIndex % numItems);
	let blockHeight = $derived(numItems * itemHeight);
	let globalOffset = $derived(blockHeight * Math.floor(scrollTop / blockHeight));
	let slice = $derived(shiftArray(sliceArray(items, startIndex, endIndex), numOverlap));

	const dummySymbol = Symbol('dummy item');

	/**
	 * @param{any[]}arr
	 * @param{number}start
	 * @param{number}end
	 */
	function sliceArray(arr, start, end) {
		arr = arr.slice(start, end);

		let expectedLength = end - start;

		// If we don't have enough items we'll fill it up with dummy entries.
		// This makes everything a lot easier, consistent and less edge-casey.
		while (arr.length < expectedLength) {
			arr.push(dummySymbol);
		}

		return arr;
	}

	/**
	 * @param{any[]}arr
	 * @param{number}count
	 */
	function shiftArray(arr, count) {
		// Could probably be optimized, but it runs on just dozens of items so relax.
		for (let i = 0; i < count; i++) {
			arr.unshift(arr.pop());
		}
		return arr;
	}
</script>

<div class="spacer" style="height: {spacerHeight}px;" tabindex="-1" on:wheel>
	{#each slice as item, index}
		<div>
			{@render children({
				buf_line: item,
				dummy: item === dummySymbol,
				y: globalOffset + (index < numOverlap ? blockHeight : 0)
			})}
		</div>
	{/each}
</div>

<style>
	.spacer {
		width: 100%;

		/* Prevent the translated items from bleeding through, causing more scrolling */
		overflow: hidden;

		/* 2021 inline-block happy fun time  */
		font-size: 0;
		line-height: 0;
	}
</style>
