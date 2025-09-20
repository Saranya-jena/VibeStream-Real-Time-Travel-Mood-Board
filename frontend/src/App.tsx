import React, { useEffect, useState } from "react";
import styled from "styled-components";

type Event = {
  user_id: string;
  lat: number;
  lon: number;
  timestamp: string;
  location?: string;
};

const Container = styled.div`
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
  font-family: system-ui;
`;

const Title = styled.h1`
  color: #2c3e50;
  border-bottom: 2px solid #3498db;
  padding-bottom: 10px;
`;

const EventCard = styled.div`
  background: white;
  border-radius: 8px;
  padding: 15px;
  margin: 10px 0;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  transition: transform 0.2s;

  &:hover {
    transform: translateY(-2px);
  }
`;

const UserName = styled.span`
  font-weight: bold;
  color: #3498db;
`;

const Location = styled.span`
  color: #27ae60;
  margin-left: 10px;
`;

const Coordinates = styled.div`
  color: #7f8c8d;
  font-size: 0.9em;
  margin-top: 5px;
`;

const Time = styled.div`
  color: #95a5a6;
  font-size: 0.8em;
  margin-top: 5px;
`;

const UserFilter = styled.div`
  margin-bottom: 20px;
`;

const Select = styled.select`
  padding: 8px;
  border-radius: 4px;
  border: 1px solid #bdc3c7;
  background: white;
  font-size: 1em;
  color: #2c3e50;

  &:focus {
    outline: none;
    border-color: #3498db;
  }
`;

function App() {
  const [events, setEvents] = useState<Event[]>([]);
  const [selectedUser, setSelectedUser] = useState<string>("");
  const [users, setUsers] = useState<Set<string>>(new Set());

  // Fetch events periodically
  useEffect(() => {
    const fetchEvents = async () => {
      try {
        const url = selectedUser
          ? `http://localhost:8082/events?user_id=${selectedUser}&limit=50`
          : "http://localhost:8082/events?limit=50";
        
        const response = await fetch(url);
        const data = await response.json();
        setEvents(data);

        // Update users list
        const newUsers = new Set<string>();
        data.forEach((event: Event) => newUsers.add(event.user_id));
        setUsers(newUsers);
      } catch (error) {
        console.error("Error fetching events:", error);
      }
    };

    // Initial fetch
    fetchEvents();

    // Set up polling every 5 seconds
    const interval = setInterval(fetchEvents, 5000);

    // Cleanup
    return () => clearInterval(interval);
  }, [selectedUser]);

  return (
    <Container>
      <Title>🌍 Garmin Mock</Title>
      
      <UserFilter>
        <Select 
          value={selectedUser} 
          onChange={(e) => setSelectedUser(e.target.value)}
        >
          <option value="">All Users</option>
          {[...users].map(user => (
            <option key={user} value={user}>{user}</option>
          ))}
        </Select>
      </UserFilter>

      <div>
        {events.map((event, index) => (
          <EventCard key={index}>
            <UserName>{event.user_id}</UserName>
            {event.location && (
              <Location>📍 {event.location}</Location>
            )}
            <Coordinates>
              📌 {event.lat.toFixed(4)}°, {event.lon.toFixed(4)}°
            </Coordinates>
            <Time>
              🕒 {new Date(event.timestamp).toLocaleTimeString()}
            </Time>
          </EventCard>
        ))}
        {events.length === 0 && (
          <EventCard>
            No events found
          </EventCard>
        )}
      </div>
    </Container>
  );
}

export default App;
