<script>
  import { selected_match } from "./store";
  /** @type {RipgrepMatch}*/
  export let match;
  /** @type {string}*/
  export let search_term;
  /** @type {string}*/
  export let replace_term;

  /** @param {RipgrepMatch} match*/
  function replace_match(match) {
    console.log("replace_match", match);
  }

  // TODO: this should really be match.Col
  // this will always highlight the first occurence
  // but that might not be the matched one
  $: start_index = match.MatchedLine.indexOf(search_term);
  $: start = match.MatchedLine.slice(0, start_index);
  $: end = match.MatchedLine.slice(start_index + search_term.length);
</script>

<button
  on:click={() => {
    replace_match(match);
  }}
  class:bg-sky={$selected_match === match}
  class="cursor-pointer snap-start flex justify-start p-2 w-full"
>
  <div class="flex items-center h-full">
    <code class="language-go">
      <div>{start}</div>
    </code>
  </div>

  <div class="flex items-center h-full whitespace-pre">
    <div
      class="font-mono px-1 spectre-match line-through bg-flamingo text-surface1 rounded-sm flex whitespace-pre-wrap"
    >
      {search_term}
    </div>
    <div
      class="ml-1 px-1 spectre-match whitespace-pre bg-surface1 text-flamingo rounded-sm"
    >
      {replace_term}
    </div>
  </div>
  <div class="flex items-center h-full">
    <code class="language-go">
      <div>{end}</div>
    </code>
  </div>
</button>

<style>
  .spectre-match {
    font-family: Consolas, Monaco, "Andale Mono", "Ubuntu Mono", monospace;
    font-size: 1em;
    text-align: left;
    white-space: pre;
    word-spacing: normal;
    word-break: normal;
    word-wrap: normal;
    -moz-tab-size: 2;
    -o-tab-size: 2;
    tab-size: 2;
    -webkit-hyphens: none;
    -moz-hyphens: none;
    -ms-hyphens: none;
    hyphens: none;
  }
</style>
