export namespace models {
	
	export class AppSettings {
	    theme: string;
	    defaultFormat: string;
	    sendingDelay: number;
	    confirmSend: boolean;
	    soundEnabled: boolean;
	    autoSaveTempls: boolean;
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
	        this.recentFiles = source["recentFiles"];
	    }
	}
	export class Contact {
	    firstName: string;
	    lastName: string;
	    email: string;
	    customFields?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Contact(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
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
	export class SendResult {
	    totalSent: number;
	    totalFailed: number;
	    logs: EmailLog[];
	    failedContacts?: Contact[];
	
	    static createFrom(source: any = {}) {
	        return new SendResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSent = source["totalSent"];
	        this.totalFailed = source["totalFailed"];
	        this.logs = this.convertValues(source["logs"], EmailLog);
	        this.failedContacts = this.convertValues(source["failedContacts"], Contact);
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

