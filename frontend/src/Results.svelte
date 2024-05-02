<script>
  import { onMount } from "svelte";
  import "./assets/prism.css";
  import ChevronDown from "./ChevronDown.svelte";
  import ChevronUp from "./ChevronUp.svelte";
  import Matches from "./Matches.svelte";
  import ResultsHeader from "./ResultsHeader.svelte";
  import { map_results } from "./results.service";

  /*** @type {RipgrepResultApi} */
  export let results;
  let collapsed = false;
  onMount(async () => {
    await import("./assets/prism.js");
    setTimeout(() => {
      window.Prism.highlightAll(false);
    });
  });

  $: mapped = map_results(results);

  $: {
    if (mapped && window.Prism) {
      window.Prism.highlightAll(false);
    }
  }
</script>

{#if results && mapped}
  {#each mapped as { path, matches }}
    {#if !path.includes(":")}
      <div class="flex pt-2 snap-start">
        <div class="text-blue">
          {#if !collapsed}
            <ChevronDown></ChevronDown>
          {:else}
            <ChevronUp></ChevronUp>
          {/if}
        </div>
        <div>
          <ResultsHeader {path} match_count={matches.length}></ResultsHeader>
          <Matches {matches}></Matches>
        </div>
      </div>
    {/if}
  {/each}
{/if}
