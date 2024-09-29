(async () => {
  const pageSource = document.documentElement.outerHTML;
  const hostname = window.location.hostname;
  const hostnameParts = hostname.split(".");
  let domain = [
    hostnameParts[hostnameParts.length - 1],
    hostnameParts[hostnameParts.length - 2],
  ];
  if (domain[1] <= 3) {
    domain.push(hostnameParts[hostnameParts.length - 3]);
  }
  console.log(domain);
  domain = btoa(domain.reverse().join("."));
  console.log(domain);

  const response = await fetch(`http://localhost:8080/check/${domain}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'text/plain'
    },
    body: pageSource
  });

})();
