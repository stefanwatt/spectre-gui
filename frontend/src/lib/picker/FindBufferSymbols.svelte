<script lang="ts">
	import { FindBufferSymbols } from '$lib/wailsjs/go/picker/LspSymbolsPicker';
	import Picker from './Picker.svelte';
	import { nestedState } from '$lib/state.svelte';

	$effect(() => {});

	function onQueryChanged(query: string) {
		FindBufferSymbols(query).then((updatedResults) => {
			nestedState.pickerResults = updatedResults;
		});
	}
</script>

<Picker {onQueryChanged} mode={'find-buffer-symbols'} title={'Find Buffer Symbols'} />
