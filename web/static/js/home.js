ready(() =>
  ((u, el) => {
    "use strict";

    const usersComp = (users) =>
      el(
        "ul",
        null,
        ...(
          users.map((user) => el("li", null, user.nickname))
        ),
      );

    const HACKER_STATE = {
      CLOSED: "Hackerspace is closed.",
      FOREVER_ALONE: "There is one person in the hackerspace.",
      PARTY: (num) => "There are " + num + " people in the hackerspace.",
    };

    const users = valoo([]);

    users((data) => {
      let usersNode = u("#users");
      usersNode.replace(usersComp(data));
      switch (data.length) {
        case 0:
          u("#online").text(HACKER_STATE.CLOSED);
          break;
        case 1:
          u("#online").text(HACKER_STATE.FOREVER_ALONE);
          break;
        default:
          u("#online").text(HACKER_STATE.PARTY(data.length));
          break;
      }
    });

    const downloadUsers = () => {
      u("#users").text("Loading...");
      fetch("/api/v1/users?online=true")
        .then((response) => response.json())
        .then((data) => users(data))
        .catch(() => {
          u("#users").text("Failed to load users data.");
        });
    };

    downloadUsers();
    window.setInterval(downloadUsers, 1000 * 60 * 2);
  })(u, el)
);
