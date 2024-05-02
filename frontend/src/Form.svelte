<script>
  export let search_term = "utils",
    dir = "/home/stefan/Projects/nvim-float";
</script>

<div class="flex">
  <input
    bind:value={search_term}
    type="text"
    placeholder="Search..."
    class="w-full input input-primary bg-base text-text border-text border rounded-md px-2 py-1"
  />
  <span class="mx-2"></span>
  <input
    bind:value={dir}
    type="text"
    placeholder="Search..."
    class="w-full input input-primary bg-base text-text border-text border rounded-md px-2 py-1"
  />
</div>
