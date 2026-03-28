<script lang="ts">
	import { FindReferences } from '@bindings/nvim-gui/features/picker/lspreferencespicker';
	import Picker from './Picker.svelte';
	import { nestedState } from '$lib/state.svelte';

	function onQueryChanged(query: string) {
		FindReferences(query).then((updatedResults) => {
			nestedState.pickerResults = updatedResults.filter((r): r is NonNullable<typeof r> => r != null);
		}).catch((err) => {
			console.error('FindReferences RPC failed', err);
		});
	}
</script>

<Picker {onQueryChanged} mode={'find-references'} title={'Find References'} />
