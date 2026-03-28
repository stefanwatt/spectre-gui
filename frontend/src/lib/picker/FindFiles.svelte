<script lang="ts">
	import { FindFiles } from '@bindings/nvim-gui/features/picker/picker.js';
	import Picker from './Picker.svelte';
	import { nestedState } from '$lib/state.svelte';

	$effect(() => {});

	function onQueryChanged(query: string) {
		FindFiles(query).then((updatedResults) => {
			nestedState.pickerResults = updatedResults;
		});
	}
</script>

<Picker {onQueryChanged} mode={'find-files'} title={'Find Files'} />
