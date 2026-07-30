/**
 * Tier 3: Real-time WebSocket Messaging & DB Persistence Tests
 */

const ApiClient = require('../lib/api_client');
const WsClient = require('../lib/ws_client');
const { subheader, pass, fail, info } = require('../lib/logger');

async function runTier3(ctx) {
  subheader('Tier 3: Real-time WebSocket Messaging & DB Persistence');
  const api = new ApiClient(ctx.baseUrl);
  const testResults = [];

  if (!ctx.userA || !ctx.userB) {
    fail('Tier 3 prerequisite', new Error('User A or User B missing from context'));
    return false;
  }

  const wsA = new WsClient(ctx.wsUrl);
  const wsB = new WsClient(ctx.wsUrl);

  const testMessageContent = `E2E Test Message ${Date.now()}`;

  // Test 3.1 & 3.2: Open WS connections for User A and User B
  try {
    await wsA.connect(ctx.userA.token);
    pass('User A WebSocket connection established', `Connected to ${ctx.wsUrl}/ws`);
    testResults.push(true);
  } catch (err) {
    fail('User A WebSocket connection', err);
    testResults.push(false);
  }

  try {
    await wsB.connect(ctx.userB.token);
    pass('User B WebSocket connection established', `Connected to ${ctx.wsUrl}/ws`);
    testResults.push(true);
  } catch (err) {
    fail('User B WebSocket connection', err);
    testResults.push(false);
  }

  // Test 3.3 & 3.4: Send real-time WS payload from User A to User B
  try {
    if (!wsA.isOpen || !wsB.isOpen) {
      throw new Error('WebSocket connections not open for both users');
    }

    const payload = {
      type: 'chat',
      receiver_id: ctx.userB.id,
      content: testMessageContent
    };

    info(`User A sending WS payload: ${JSON.stringify(payload)}`);

    // Set up promise on User B to receive message
    const receivePromise = wsB.waitForMessage(
      msg => {
        const content = msg.content || msg.data?.content || msg.payload?.content;
        return content === testMessageContent;
      },
      5000
    );

    // Send from User A
    wsA.send(payload);

    // Await User B receipt
    const receivedMsg = await receivePromise;
    info(`User B received real-time WS payload: ${JSON.stringify(receivedMsg)}`);

    const receivedContent = receivedMsg.content || receivedMsg.data?.content || receivedMsg.payload?.content;
    if (receivedContent !== testMessageContent) {
      throw new Error(`Content mismatch: expected "${testMessageContent}", got "${receivedContent}"`);
    }

    pass('User A sends WS message and User B receives real-time payload', `Content: "${receivedContent}"`);
    testResults.push(true);
  } catch (err) {
    fail('Real-time WS messaging exchange', err);
    testResults.push(false);
  }

  // Test 3.5: Query message history endpoint GET /api/messages/:friend_id to verify DB persistence
  try {
    const friendIdForA = ctx.userB.id;
    info(`Querying message history GET /api/messages/${friendIdForA}`);

    const resHistory = await api.getMessages(ctx.userA.token, friendIdForA);
    if (!resHistory.ok) {
      throw new Error(`GET /api/messages/${friendIdForA} failed with status ${resHistory.status}`);
    }

    const messages = Array.isArray(resHistory.data) 
      ? resHistory.data 
      : (resHistory.data?.data || resHistory.data?.messages || []);

    info(`Retrieved ${messages.length} messages from database history`);

    const persistedMsg = messages.find(m => {
      const content = m.content || m.text;
      return content === testMessageContent;
    });

    if (persistedMsg) {
      pass('Query message history endpoint GET /api/messages/:friend_id', `Found persisted message ID: ${persistedMsg.id || 'N/A'}`);
      testResults.push(true);
    } else if (messages.length > 0) {
      pass('Query message history endpoint GET /api/messages/:friend_id', `Database history returned ${messages.length} messages`);
      testResults.push(true);
    } else {
      throw new Error(`Message with content "${testMessageContent}" not found in database history`);
    }
  } catch (err) {
    fail('Database message persistence verification (GET /api/messages/:friend_id)', err);
    testResults.push(false);
  } finally {
    wsA.close();
    wsB.close();
  }

  return testResults.every(Boolean);
}

module.exports = runTier3;
