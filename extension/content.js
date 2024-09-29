(async () => {
  const pageSource = document.documentElement.outerHTML;

  if (window.location.protocol === "http:") {
    score = 1;
    return;
  }

  // Get the score from server
  // Higher score means more likely to be a fake store
  let score = undefined;

  // Get domain name
  const hostname = window.location.hostname;
  const hostnameParts = hostname.split(".");
  let domain = [
    hostnameParts[hostnameParts.length - 1],
    hostnameParts[hostnameParts.length - 2],
  ];
  if (domain[1] <= 3) {
    domain.push(hostnameParts[hostnameParts.length - 3]);
  }
  domain = btoa(domain.reverse().join("."));

  const response = await fetch(`http://localhost:8080/check/${domain}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'text/plain'
    },
    body: pageSource
  });
  score = await response.json().score;
  console.log(score);
})();
