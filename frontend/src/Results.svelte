<script>
  import { onMount } from "svelte";
  import "./assets/prism.css";
  import ChevronDown from "./ChevronDown.svelte";
  import ChevronUp from "./ChevronUp.svelte";
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
        <div class="grid grid-cols-[1fr,15fr] pt-2 snap-start">
          <div class="text-blue flex justify-end items-center w-full">
            {#if !collapsed}
              <ChevronDown></ChevronDown>
            {:else}
              <ChevronUp></ChevronUp>
            {/if}
          </div>
          <ResultsHeader {path} match_count={matches.length}></ResultsHeader>
          {#each matches as match}
            {#if match.Path.length < 15 && match.Col.length < 5}
              <div
                class="pr-2 text-overlay0 flex justify-end items-center w-full"
              >
                {match.Row}:{match.Col}
              </div>
              <Match {match}></Match>
            {/if}
          {/each}
        </div>
      {/if}
    {/each}
  </div>
{/if}
