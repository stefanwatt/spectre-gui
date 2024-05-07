<script>
  import { onMount } from "svelte";
  import "./assets/prism.css";
  import ChevronDown from "./ChevronDown.svelte";
  import ChevronUp from "./ChevronUp.svelte";
  import ResultsHeader from "./ResultsHeader.svelte";
  import { highlight_all } from "./results.service";
  import { results } from "./store";
  import Match from "./Match.svelte";
  import NoResults from "./NoResults.svelte";

  let collapsed = false;
  onMount(async () => {
    await import("./assets/prism.js");
    setTimeout(() => {
      highlight_all();
    });
  });
</script>

<svelte:head>
  <script
    src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.27.0/plugins/filter-highlight-all/prism-filter-highlight-all.min.js"
  ></script>
</svelte:head>
<div class="flex flex-col w-full">
  {#each $results as item(item.path)}
    <div class="grid grid-cols-[1fr,15fr] pt-2 snap-start">
      <div class="mb-2 text-blue flex justify-end items-center w-full">
        {#if !collapsed}
          <ChevronDown></ChevronDown>
        {:else}
          <ChevronUp></ChevronUp>
        {/if}
      </div>
      <div class="mb-2">
        <ResultsHeader path ={item.path} match_count={item.matches.length}></ResultsHeader>
      </div>
      {#each item.matches as match (match.Id)}
        {#if match.Col < 10000}
          <div class="pr-2 text-overlay0 flex justify-end items-center w-full">
            {match.Row}:{match.Col}
          </div>
          <Match {match}></Match>
        {/if}
      {/each}
    </div>
  {:else}
    <NoResults></NoResults>
  {/each}
</div>
