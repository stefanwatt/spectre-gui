<script>
  import { onMount } from "svelte";
  import "./assets/prism.css";
  import ChevronDown from "./ChevronDown.svelte";
  import ChevronUp from "./ChevronUp.svelte";
  import Matches from "./Matches.svelte";
  import ResultsHeader from "./ResultsHeader.svelte";
  import { map_results, highlight_all } from "./results.service";
  import Match from "./Match.svelte";

  /*** @type {RipgrepResultApi} */
  export let results;
  let collapsed = false;
  onMount(async () => {
    await import("./assets/prism.js");
    setTimeout(() => {
      highlight_all();
    });
  });

  $: mapped = map_results(results);

  $: {
    if (mapped) {
      highlight_all();
    }
  }
</script>

{#if results && mapped}
  <div class="flex flex-col w-full">
    {#each mapped as { path, matches }}
      {#if !path.includes(":")}
        <div class="grid grid-cols-[auto,1fr] pt-2 snap-start">
          <div class="text-blue">
            {#if !collapsed}
              <ChevronDown></ChevronDown>
            {:else}
              <ChevronUp></ChevronUp>
            {/if}
          </div>
          <div>
            <ResultsHeader {path} match_count={matches.length}></ResultsHeader>

            {#each matches as match}
              {#if match.Path.length < 15 && match.Col.length < 5}
                <Match {match}></Match>
              {/if}
            {/each}
          </div>
        </div>
      {/if}
    {/each}
  </div>
{/if}
