<script>
  import { selected_match, search_term, replace_term } from "./store";
  /** @type {RipgrepMatch}*/
  export let match;
  /** @param {RipgrepMatch} match*/
  function replace_match(match) {
    console.log("replace_match", match);
  }
</script>

<button
  on:click={() => {
    replace_match(match);
  }}
  class:bg-surface1={$selected_match === match}
  class="cursor-pointer rounded-sm snap-start flex justify-start p-1 m-1 w-full"
>
  <div class="flex items-center h-full">
    <code class="language-go">
      <div>{match.TextBeforeMatch}</div>
    </code>
  </div>

  <div class="flex items-center h-full whitespace-pre">
    <div
      class="font-mono px-1 spectre-match line-through bg-flamingo text-surface1 rounded-sm flex whitespace-pre-wrap"
    >
      {match.MatchedText}
    </div>
    <div
      class="ml-1 px-1 spectre-match whitespace-pre bg-surface1 text-flamingo rounded-sm"
    >
      {match.ReplacementText}
    </div>
  </div>
  <div class="flex items-center h-full">
    <code class="language-go bg-transparent">
      <div>{match.TextAfterMatch}</div>
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
