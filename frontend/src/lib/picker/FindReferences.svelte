<script lang="ts">
	import { FindReferences } from '@bindings/nvim-gui/features/picker/lspreferencespicker.js';
	import Picker from './Picker.svelte';
	import { nestedState } from '$lib/state.svelte';

	$effect(() => {});

	function onQueryChanged(query: string) {
		FindReferences(query).then((updatedResults) => {
			nestedState.pickerResults = updatedResults;
		});
	}
</script>

<Picker {onQueryChanged} mode={'find-references'} title={'Find References'} />
