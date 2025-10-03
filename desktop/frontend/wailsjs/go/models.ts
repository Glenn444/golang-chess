export namespace board {
	
	export class Square {
	    Occupied: boolean;
	    Piece: any;
	
	    static createFrom(source: any = {}) {
	        return new Square(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Occupied = source["Occupied"];
	        this.Piece = source["Piece"];
	    }
	}
	export class GameState {
	    CurrentPlayer: string;
	    Board: Square[][];
	
	    static createFrom(source: any = {}) {
	        return new GameState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CurrentPlayer = source["CurrentPlayer"];
	        this.Board = this.convertValues(source["Board"], Square);
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

