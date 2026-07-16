export namespace campaign {
	
	export class Attempt {
	    number: number;
	    status: string;
	    error?: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new Attempt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.number = source["number"];
	        this.status = source["status"];
	        this.error = source["error"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SenderStatus {
	    available: boolean;
	    state: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SenderStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.state = source["state"];
	        this.message = source["message"];
	    }
	}
	export class SenderCapabilities {
	    supportsHTML: boolean;
	    supportsAttachments: boolean;
	    supportsMultipleAccounts: boolean;
	    supportsSharedMailbox: boolean;
	    supportsDraftOnly: boolean;
	    supportsScheduling: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SenderCapabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supportsHTML = source["supportsHTML"];
	        this.supportsAttachments = source["supportsAttachments"];
	        this.supportsMultipleAccounts = source["supportsMultipleAccounts"];
	        this.supportsSharedMailbox = source["supportsSharedMailbox"];
	        this.supportsDraftOnly = source["supportsDraftOnly"];
	        this.supportsScheduling = source["supportsScheduling"];
	    }
	}
	export class PreflightIssue {
	    severity: string;
	    code: string;
	    message: string;
	    contactIndex?: number;
	    email?: string;
	    fieldId?: string;
	    count?: number;
	
	    static createFrom(source: any = {}) {
	        return new PreflightIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.severity = source["severity"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.contactIndex = source["contactIndex"];
	        this.email = source["email"];
	        this.fieldId = source["fieldId"];
	        this.count = source["count"];
	    }
	}
	export class PreflightResult {
	    canSend: boolean;
	    errors: PreflightIssue[];
	    warnings: PreflightIssue[];
	    recipientCount: number;
	    suppressedCount: number;
	    duplicateDropped: number;
	    duplicateGroups?: email.Group[];
	    estimatedAttempts: number;
	    estimatedDuration: number;
	    capabilities: SenderCapabilities;
	    senderStatus: SenderStatus;
	    recipients: number[];
	
	    static createFrom(source: any = {}) {
	        return new PreflightResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.canSend = source["canSend"];
	        this.errors = this.convertValues(source["errors"], PreflightIssue);
	        this.warnings = this.convertValues(source["warnings"], PreflightIssue);
	        this.recipientCount = source["recipientCount"];
	        this.suppressedCount = source["suppressedCount"];
	        this.duplicateDropped = source["duplicateDropped"];
	        this.duplicateGroups = this.convertValues(source["duplicateGroups"], email.Group);
	        this.estimatedAttempts = source["estimatedAttempts"];
	        this.estimatedDuration = source["estimatedDuration"];
	        this.capabilities = this.convertValues(source["capabilities"], SenderCapabilities);
	        this.senderStatus = this.convertValues(source["senderStatus"], SenderStatus);
	        this.recipients = source["recipients"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RecipientResult {
	    contactIndex: number;
	    contactId?: string;
	    email: string;
	    firstName?: string;
	    lastName?: string;
	    status: string;
	    attempts: Attempt[];
	
	    static createFrom(source: any = {}) {
	        return new RecipientResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contactIndex = source["contactIndex"];
	        this.contactId = source["contactId"];
	        this.email = source["email"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.status = source["status"];
	        this.attempts = this.convertValues(source["attempts"], Attempt);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CampaignResult {
	    state: string;
	    attempted: number;
	    submitted: number;
	    failed: number;
	    skipped: number;
	    cancelled: number;
	    fatalError?: string;
	    recipientResults: RecipientResult[];
	    preflight?: PreflightResult;
	
	    static createFrom(source: any = {}) {
	        return new CampaignResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.attempted = source["attempted"];
	        this.submitted = source["submitted"];
	        this.failed = source["failed"];
	        this.skipped = source["skipped"];
	        this.cancelled = source["cancelled"];
	        this.fatalError = source["fatalError"];
	        this.recipientResults = this.convertValues(source["recipientResults"], RecipientResult);
	        this.preflight = this.convertValues(source["preflight"], PreflightResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	

}

export namespace email {
	
	export class Group {
	    key: string;
	    indices: number[];
	
	    static createFrom(source: any = {}) {
	        return new Group(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.indices = source["indices"];
	    }
	}

}

export namespace models {
	
	export class AppSettings {
	    theme: string;
	    defaultFormat: string;
	    sendingDelay: number;
	    confirmSend: boolean;
	    soundEnabled: boolean;
	    autoSaveTempls: boolean;
	    duplicatePolicy: string;
	    schemaVersion: number;
	    recentFiles: string[];
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.defaultFormat = source["defaultFormat"];
	        this.sendingDelay = source["sendingDelay"];
	        this.confirmSend = source["confirmSend"];
	        this.soundEnabled = source["soundEnabled"];
	        this.autoSaveTempls = source["autoSaveTempls"];
	        this.duplicatePolicy = source["duplicatePolicy"];
	        this.schemaVersion = source["schemaVersion"];
	        this.recentFiles = source["recentFiles"];
	    }
	}
	export class Contact {
	    id: string;
	    firstName: string;
	    lastName: string;
	    email: string;
	    customFields?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Contact(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.email = source["email"];
	        this.customFields = source["customFields"];
	    }
	}
	export class EmailLog {
	    firstName: string;
	    lastName: string;
	    email: string;
	    status: string;
	    errorMessage?: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new EmailLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.email = source["email"];
	        this.status = source["status"];
	        this.errorMessage = source["errorMessage"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EmailRequest {
	    contacts: Contact[];
	    subjectTemplate: string;
	    bodyTemplate: string;
	    isHTML: boolean;
	    attachments: string[];
	    cc?: string;
	    bcc?: string;
	    ccTemplate?: string;
	    bccTemplate?: string;
	
	    static createFrom(source: any = {}) {
	        return new EmailRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contacts = this.convertValues(source["contacts"], Contact);
	        this.subjectTemplate = source["subjectTemplate"];
	        this.bodyTemplate = source["bodyTemplate"];
	        this.isHTML = source["isHTML"];
	        this.attachments = source["attachments"];
	        this.cc = source["cc"];
	        this.bcc = source["bcc"];
	        this.ccTemplate = source["ccTemplate"];
	        this.bccTemplate = source["bccTemplate"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EmailTemplate {
	    id: string;
	    name: string;
	    subject: string;
	    body: string;
	    isHTML: boolean;
	    isBuiltIn: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new EmailTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.subject = source["subject"];
	        this.body = source["body"];
	        this.isHTML = source["isHTML"];
	        this.isBuiltIn = source["isBuiltIn"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileInfo {
	    name: string;
	    path: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	    }
	}
	export class ParseResult {
	    contacts: Contact[];
	    errors?: string[];
	    warnings?: string[];
	    total: number;
	    headers?: string[];
	    duplicates?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ParseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contacts = this.convertValues(source["contacts"], Contact);
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	        this.total = source["total"];
	        this.headers = source["headers"];
	        this.duplicates = source["duplicates"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TestEmailRequest {
	    testAddress: string;
	    subjectTemplate: string;
	    bodyTemplate: string;
	    isHTML: boolean;
	    attachments: string[];
	    cc?: string;
	    bcc?: string;
	    ccTemplate?: string;
	    bccTemplate?: string;
	    sampleFirstName: string;
	    sampleLastName: string;
	    contact: Contact;
	    overwriteEmail: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TestEmailRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.testAddress = source["testAddress"];
	        this.subjectTemplate = source["subjectTemplate"];
	        this.bodyTemplate = source["bodyTemplate"];
	        this.isHTML = source["isHTML"];
	        this.attachments = source["attachments"];
	        this.cc = source["cc"];
	        this.bcc = source["bcc"];
	        this.ccTemplate = source["ccTemplate"];
	        this.bccTemplate = source["bccTemplate"];
	        this.sampleFirstName = source["sampleFirstName"];
	        this.sampleLastName = source["sampleLastName"];
	        this.contact = this.convertValues(source["contact"], Contact);
	        this.overwriteEmail = source["overwriteEmail"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace services {
	
	export class MergeResult {
	    subject: string;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new MergeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.subject = source["subject"];
	        this.body = source["body"];
	    }
	}

}

