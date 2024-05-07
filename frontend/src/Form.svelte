<script>
  import { debounce } from "./utils.service.js";
  import { search_term, replace_term, dir, include, exclude } from "./store.js";

  /**
   * @param {KeyboardEvent & { target: HTMLInputElement }} e
   */
  async function debounced_search_term(e) {
    $search_term = await debounce(e.target.value);
  }

  /**
   * @param {KeyboardEvent & { target: HTMLInputElement }} e
   */
  async function debounced_dir(e) {
    $dir = await debounce(e.target.value);
  }

  /**
   * @param {KeyboardEvent & { target: HTMLInputElement }} e
   */
  async function debounced_exclude(e) {
    $exclude = await debounce(e.target.value);
  }

  /**
   * @param {KeyboardEvent & { target: HTMLInputElement }} e
   */
  async function debounced_include({ target: { value } }) {
    $include = await debounce(value);
  }
</script>

<div class="flex">
  <div class="w-full">
    <input
      autofocus
      on:keyup={debounced_search_term}
      value={$search_term}
      type="text"
      placeholder="Search..."
      class="w-full input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
    />
  </div>
  <div class="w-full pl-2">
    <input
      bind:value={$replace_term}
      type="text"
      placeholder="Replace..."
      class="w-full input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
    />
  </div>
</div>
<div class="py-2 flex">
  <div class="w-full">
    <input
      on:keyup={debounced_dir}
      value={$dir}
      type="text"
      placeholder="Search..."
      class="w-full input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
    />
  </div>
  <div class="w-full flex">
    <input
      on:keyup={debounced_exclude}
      value={$exclude}
      type="text"
      placeholder="eg *service.go,src/**/exclude"
      class="w-full ml-2 input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
    />
    <input
      on:keyup={debounced_include}
      value={$include}
      type="text"
      placeholder="eg *service.go,src/**/include"
      class="w-full ml-2 input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
    />
  </div>
</div>
