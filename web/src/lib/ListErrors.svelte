<script>
	export let errors;
</script>

{#if errors}
  <div class="alert alert-danger alert-dismissible fade show" role="alert">
    <strong>Failed!</strong> Please fix errors marked below.
    <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    <ul class="error-messages">
      {#each Object.keys(errors) as key}
        <li>{key} {errors[key]}</li>
      {/each}
    </ul>
  </div>
{/if}
