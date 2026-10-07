// Scope WebKit's network-abort page error to the deliberate recovery fault.
// Each page owns its registry; only an actual route.abort registers a URL.
class ExpectedAbort {
  constructor(){this.active=false;this.messages=new Set();this.consoleUrls=new Set();}
  begin(){this.active=true;}
  record(url){
    if(!this.active) throw Error('Abort outside recovery fault phase');
    this.consoleUrls.add(url);
    // WebKit 26.6 reports this URL with the scheme and first slash omitted.
    this.messages.add(`${url.replace(/^https?:\//,'')} due to access control checks.`);
  }
  consume(message){
    if(!this.active || !this.messages.has(message)) return false;
    this.messages.delete(message);return true;
  }
  consumeConsole(message,url){
    if(!this.active || message!=='Failed to load resource: net::ERR_FAILED' || !this.consoleUrls.has(url)) return false;
    this.consoleUrls.delete(url);return true;
  }
  end(){this.active=false;this.messages.clear();this.consoleUrls.clear();}
}
function isBrowserError(type,message){return type==='error' || (type==='log' && /Exception|Error|Invalid argument|overflowed/.test(message));}
function isInitialSessionChallenge(message,url,sessionUrl,status){
 return status===401 && url===sessionUrl && message==='Failed to load resource: the server responded with a status of 401 (Unauthorized)';
}
module.exports={ExpectedAbort,isBrowserError,isInitialSessionChallenge};
